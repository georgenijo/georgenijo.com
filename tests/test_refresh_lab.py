"""Network-free tests for event/manual George Lab refresh behavior."""

import importlib.util
import http.client
import unittest
from pathlib import Path
from unittest import mock


SCRIPT = Path(__file__).resolve().parents[1] / "scripts" / "refresh-lab.py"
SPEC = importlib.util.spec_from_file_location("refresh_lab", SCRIPT)
refresh_lab = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(refresh_lab)


class RefreshLabTests(unittest.TestCase):
    def test_manual_refresh_accepts_no_project(self):
        refresh_lab.validate_project(None)

    def test_dispatch_accepts_every_documented_project(self):
        for project in refresh_lab.PROJECTS:
            refresh_lab.validate_project(project["repo"])

    def test_dispatch_rejects_unknown_project(self):
        with self.assertRaisesRegex(ValueError, "unknown project"):
            refresh_lab.validate_project("not-a-project")

    def test_snapshot_is_deterministic_with_mocked_github_api(self):
        def fake_request(path, token):
            self.assertEqual(token, "test-token")
            if path.endswith("/releases/latest"):
                return {
                    "tag_name": "v1.2.3",
                    "name": "",
                    "html_url": "https://example.invalid/release",
                    "published_at": "2026-07-24T12:00:00Z",
                }
            if path.startswith("/search/issues"):
                return {"total_count": 2}
            repo = path.rsplit("/", 1)[-1]
            return {
                "html_url": f"https://github.com/georgenijo/{repo}",
                "pushed_at": "2026-07-24T11:00:00Z",
            }

        with mock.patch.object(refresh_lab, "request_json", side_effect=fake_request):
            snapshot = refresh_lab.build_snapshot(
                "test-token", generated_at="2026-07-24T13:00:00Z"
            )

        self.assertEqual(snapshot["schemaVersion"], 1)
        self.assertEqual(snapshot["generatedAt"], "2026-07-24T13:00:00Z")
        self.assertEqual(
            [project["id"] for project in snapshot["projects"]],
            [project["repo"] for project in refresh_lab.PROJECTS],
        )
        self.assertTrue(all(project["metadataAvailable"] for project in snapshot["projects"]))
        self.assertTrue(all(project["openIssues"] == 2 for project in snapshot["projects"]))

    def test_github_request_retries_transient_disconnect(self):
        response = mock.MagicMock()
        response.__enter__.return_value = ["ok"]
        with mock.patch.object(
            refresh_lab.urllib.request,
            "urlopen",
            side_effect=[http.client.RemoteDisconnected(), response],
        ) as urlopen, mock.patch.object(refresh_lab.json, "load", return_value={"ok": True}), mock.patch.object(
            refresh_lab.time, "sleep"
        ) as sleep:
            result = refresh_lab.request_json("/test", None)

        self.assertEqual(result, {"ok": True})
        self.assertEqual(urlopen.call_count, 2)
        sleep.assert_called_once_with(1)


if __name__ == "__main__":
    unittest.main()
