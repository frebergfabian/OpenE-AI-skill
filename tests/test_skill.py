import importlib.util
import tempfile
import unittest
import zipfile
from pathlib import Path

spec = importlib.util.spec_from_file_location("skill", Path(__file__).resolve().parents[1] / "tools/skill.py")
skill = importlib.util.module_from_spec(spec)
spec.loader.exec_module(skill)


class SkillTests(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory()
        self.addCleanup(self.temporary.cleanup)
        self.base = Path(self.temporary.name)
        self.root = self.base / "opene-platform"
        self.root.mkdir()
        self.entry = self.root / "SKILL.md"
        self.entry.write_text("---\nname: opene-platform\ndescription: Test skill\n---\n")
        assets = self.root / "assets"
        assets.mkdir()
        self.fixture = assets / "test.oeflow"
        self.fixture.write_text("<Flow><enabled>false</enabled></Flow>")

    def test_package_roundtrip_and_runtime_exclusions(self):
        (self.root / ".env").write_text("secret")
        (self.root / "payload.zip").write_bytes(b"runtime")
        output = self.base / "test.skill"
        skill.package(self.root, output)
        with zipfile.ZipFile(output) as archive:
            self.assertEqual(set(archive.namelist()), {
                "opene-platform/SKILL.md", "opene-platform/assets/test.oeflow"})
            extracted = self.base / "extracted"
            archive.extractall(extracted)
        skill.validate(extracted / "opene-platform")
        second = self.base / "second.skill"
        skill.package(self.root, second)
        self.assertEqual(output.read_bytes(), second.read_bytes())

    def test_runtime_symlinks_excluded_but_skill_symlinks_rejected(self):
        cache = self.root / "node_modules" / ".bin"
        cache.mkdir(parents=True)
        try:
            (cache / "playwright").symlink_to(self.entry)
        except OSError:
            self.skipTest("Symlink creation is unavailable")
        self.assertEqual(len(skill.validate(self.root)), 2)
        (self.root / "reference.md").symlink_to(self.entry)
        with self.assertRaises(ValueError):
            skill.validate(self.root)

    def test_broken_link_and_escape_are_rejected(self):
        for target in ("missing.md", "../outside.md"):
            (self.base / "outside.md").write_text("private")
            self.entry.write_text("---\nname: opene-platform\ndescription: Test skill\n---\n"
                                  f"[Reference]({target})")
            with self.assertRaises(ValueError):
                skill.validate(self.root)

    def test_active_and_entity_fixtures_are_rejected(self):
        for xml in ("<Flow><enabled>true</enabled></Flow>",
                    '<!DOCTYPE Flow [<!ENTITY value "false">]><Flow><enabled>&value;</enabled></Flow>'):
            self.fixture.write_text(xml)
            with self.assertRaises(ValueError):
                skill.validate(self.root)

    def test_conflict_marker_is_rejected(self):
        with self.entry.open("a") as handle:
            handle.write("<<<<<<< HEAD\n")
        with self.assertRaises(ValueError):
            skill.validate(self.root)


if __name__ == "__main__":
    unittest.main()
