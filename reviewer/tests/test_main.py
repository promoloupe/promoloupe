import pytest

from promoloupe_reviewer import main


def test_main_prints_greeting(capsys: pytest.CaptureFixture[str]) -> None:
    main()

    assert capsys.readouterr().out == "Hello from promoloupe-reviewer!\n"
