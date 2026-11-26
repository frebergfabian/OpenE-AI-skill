#!/usr/bin/env python3
"""Reconcile the build report with enabled OpenHierarchy module class names.

Usage: python report-activation.py addon-report.json jar-directory enabled-classes.txt output.json
The text file contains one enabled foreground/background classname per line,
queried from the isolated test instance. A registered class is not proof that
every class in its project is enabled or that its integration was tested.
"""
import json
import pathlib
import sys
import zipfile

if len(sys.argv) != 5:
    raise SystemExit(__doc__)
report_path, jar_dir, enabled_path, output_path = map(pathlib.Path, sys.argv[1:])
enabled = set(enabled_path.read_text(encoding="utf-8").splitlines())
report = json.loads(report_path.read_text(encoding="utf-8"))
for item in report:
    jar = jar_dir / (item["project"] + ".jar")
    if not jar.exists() or item["status"] == "built-excluded-duplicate-classes":
        item["enabledClasses"] = []
        continue
    with zipfile.ZipFile(jar) as archive:
        classes = {name[:-6].replace("/", ".") for name in archive.namelist()
                   if name.endswith(".class") and "$" not in name}
    item["enabledClasses"] = sorted(classes & enabled)
    if item["enabledClasses"]:
        item["status"] = ("built-partial-project-with-enabled-modules"
                          if item["status"].startswith("built-partial")
                          else "built-with-enabled-modules")
        item["reason"] = "At least one class registered as enabled; other classes and behavior require separate verification"
output_path.write_text(json.dumps(report, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
