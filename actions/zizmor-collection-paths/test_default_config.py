import unittest
from pathlib import Path

HERE = Path(__file__).resolve().parent
DEFAULT_CONFIG = HERE / "default-zizmor.yml"
LEGACY_CONFIG = HERE.parents[1] / ".github" / "zizmor.yml"


def _without_comments(path: Path) -> list[str]:
    lines = (line.rstrip() for line in path.read_text().splitlines())
    return [line for line in lines if line and not line.lstrip().startswith("#")]


class DefaultConfigTests(unittest.TestCase):
    def test_matches_legacy_config(self) -> None:
        # Older reusable-zizmor versions still download .github/zizmor.yml.
        self.assertEqual(_without_comments(DEFAULT_CONFIG), _without_comments(LEGACY_CONFIG))


if __name__ == "__main__":
    unittest.main()
