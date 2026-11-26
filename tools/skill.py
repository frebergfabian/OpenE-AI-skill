#!/usr/bin/env python3
"""Validate and package the repository's installable Open ePlatform skill."""
import argparse
import os
import re
import subprocess
import sys
import tempfile
import zipfile
from pathlib import Path
from xml.parsers import expat
from xml.etree import ElementTree as ET

ROOT = Path(__file__).resolve().parents[1]
SKILL = ROOT / "opene-platform"
EXCLUDED = {"node_modules", "__pycache__", "report", "results", "cache"}
EXCLUDED_SUFFIXES = {".exe", ".pyc", ".test"}
LINK = re.compile(r"(?<!!)\[[^\]]*\]\(([^)]+)\)|!\[[^\]]*\]\(([^)]+)\)")


def files(directory):
    for path in sorted(directory.rglob("*")):
        relative = path.relative_to(directory)
        if path.is_symlink():
            raise ValueError(f"Symlink in skill: {relative}")
        if path.is_file() and not EXCLUDED.intersection(relative.parts):
            if path.suffix not in EXCLUDED_SUFFIXES and path.name not in {"payload.zip", ".env"}:
                yield path


def validate(directory):
    paths = list(files(directory))
    entry = (directory / "SKILL.md").read_text(encoding="utf-8")
    if not entry.startswith("---\n") or "\n---\n" not in entry[4:]:
        raise ValueError("SKILL.md requires YAML frontmatter")
    frontmatter = entry.split("---", 2)[1]
    for field in ("name", "description"):
        if len(re.findall(rf"^{field}:\s*\S.+$", frontmatter, re.M)) != 1:
            raise ValueError(f"Expected exactly one {field}")
    for path in paths:
        if path.suffix not in {".md", ".oeflow"}:
            continue
        text = path.read_text(encoding="utf-8")
        if re.search(r"^(<<<<<<< |=======$|>>>>>>> )", text, re.M):
            raise ValueError(f"Unresolved merge in {path.name}")
        if path.suffix == ".md":
            for match in LINK.finditer(text):
                target = (match.group(1) or match.group(2)).split("#", 1)[0]
                if not target or re.match(r"[a-zA-Z][a-zA-Z0-9+.-]*:", target):
                    continue
                resolved = (path.parent / target).resolve()
                if not resolved.is_relative_to(directory.resolve()) or not resolved.exists():
                    raise ValueError(f"Broken or external local link in {path.name}: {target}")
        else:
            parser = expat.ParserCreate()

            def reject(*_):
                raise ValueError(f"DTD/entities forbidden in {path.name}")

            parser.StartDoctypeDeclHandler = reject
            parser.EntityDeclHandler = reject
            parser.ExternalEntityRefHandler = reject
            parser.Parse(text.encode("utf-8"), True)
            flow = ET.fromstring(text)
            if flow.tag != "Flow" or flow.findtext("enabled") != "false":
                raise ValueError(f"Expected an inactive Flow fixture: {path.name}")
    return paths


def package(directory, output):
    paths = validate(directory)
    if output.resolve().is_relative_to(directory.resolve()):
        raise ValueError("Place the generated archive outside the skill directory")
    output.parent.mkdir(parents=True, exist_ok=True)
    with zipfile.ZipFile(output, "w", zipfile.ZIP_DEFLATED) as archive:
        for path in paths:
            name = "opene-platform/" + path.relative_to(directory).as_posix()
            info = zipfile.ZipInfo(name, date_time=(2020, 1, 1, 0, 0, 0))
            info.compress_type = zipfile.ZIP_DEFLATED
            mode = 0o755 if path.stat().st_mode & 0o111 else 0o644
            info.external_attr = (0o100000 | mode) << 16
            archive.writestr(info, path.read_bytes())
    return len(paths)


def runtime_tests():
    subprocess.run(["node", "--test", str(SKILL / "sandbox/build/sanitize-wsdl.test.mjs")], check=True)
    with tempfile.TemporaryDirectory(prefix="opene-skill-tests-") as temporary:
        stage = Path(temporary)
        for source in (SKILL / "sandbox/portable").glob("*"):
            if source.suffix == ".go" or source.name == "go.mod":
                (stage / source.name).write_bytes(source.read_bytes())
        # Tests exercise Go helpers and simulators. No appserver payload is distributed.
        with zipfile.ZipFile(stage / "payload.zip", "w") as archive:
            archive.writestr("TEST-ONLY", "Synthetic test payload")
        subprocess.run(["go", "test", "-v", "./..."], cwd=stage, check=True)
        windows_env = {**os.environ, "GOOS": "windows", "GOARCH": "amd64", "CGO_ENABLED": "0"}
        subprocess.run(["go", "test", "-c", "-o", "windows-tests.exe", "."],
                       cwd=stage, env=windows_env, check=True)
        print("Windows x64 test binary compiled; execution requires Windows", flush=True)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output", type=Path, help="Create an installable .skill ZIP")
    parser.add_argument("--runtime-tests", action="store_true", help="Also run Go and Node tests")
    args = parser.parse_args()
    paths = validate(SKILL)
    print(f"Validated {len(paths)} skill files", flush=True)
    if args.runtime_tests:
        runtime_tests()
    if args.output:
        print(f"Packaged {package(SKILL, args.output)} files into {args.output}")


if __name__ == "__main__":
    try:
        main()
    except (ValueError, OSError, subprocess.CalledProcessError) as error:
        print(error, file=sys.stderr)
        sys.exit(1)
