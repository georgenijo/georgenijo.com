"""Regression checks for partial Fleet failures and snapshot preservation."""

import importlib.util
import json
import subprocess
import unittest
from pathlib import Path
from unittest import mock


SCRIPT = Path(__file__).resolve().parents[1] / "scripts" / "collect-burn.py"
SPEC = importlib.util.spec_from_file_location("collect_burn", SCRIPT)
collector = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(collector)


class CollectBurnTests(unittest.TestCase):
    def test_partial_fleet_failure_keeps_successful_usage(self):
        payload = json.dumps([
            {"node": "worker", "ok": True, "stdout": json.dumps({
                "totals": {"totalTokens": 120},
                "daily": [{"date": "2026-10-28", "totalTokens": 120,
                           "modelBreakdowns": [{"modelName": "gpt-demo", "totalTokens": 120}]}],
            })},
            {"node": "monitor", "ok": False, "error": "npx unavailable"},
        ])
        with mock.patch.object(collector.subprocess, "check_output", side_effect=
                               subprocess.CalledProcessError(127, "fleet", output=payload)), \
             mock.patch.object(collector, "collect_commits_local", return_value={}):
            snapshot = collector.collect_fleet_and_merge()
        self.assertEqual(snapshot["totalTokens"], 120)
        self.assertEqual(snapshot["byModelTop"][0]["tokens"], 120)
        self.assertEqual(snapshot["nodesCollected"], 1)
        self.assertEqual(snapshot["nodesSkipped"], 1)
        self.assertEqual(len(snapshot["last30"]), 30)

    def test_all_failed_nodes_return_no_snapshot(self):
        with mock.patch.object(collector, "run", return_value='[{"ok": false}]'):
            self.assertIsNone(collector.collect_fleet_and_merge())

    def test_failed_fleet_refresh_does_not_write_or_fall_back_to_local(self):
        with mock.patch("sys.argv", [str(SCRIPT), "--fleet"]), \
             mock.patch.object(collector.os, "chdir"), \
             mock.patch.object(collector, "collect_commits_local", return_value={}), \
             mock.patch.object(collector, "collect_fleet_and_merge", return_value=None), \
             mock.patch.object(collector, "build_burn_from_local") as local, \
             mock.patch("builtins.open") as write:
            self.assertEqual(collector.main(), 1)
        local.assert_not_called()
        write.assert_not_called()


if __name__ == "__main__":
    unittest.main()
