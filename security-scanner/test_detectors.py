import unittest
from detectors.missing_replay_protection import MissingReplayProtection
from detectors.missing_source_validation import MissingSourceValidation
from detectors.unrestricted_message_sender import UnrestrictedMessageSender

class TestDetectors(unittest.TestCase):
    def test_missing_replay_protection_metadata(self):
        self.assertEqual(MissingReplayProtection.ARGUMENT, "missing-replay-protection")
        self.assertTrue(len(MissingReplayProtection.HELP) > 0)
        self.assertTrue(len(MissingReplayProtection.WIKI) > 0)

    def test_missing_source_validation_metadata(self):
        self.assertEqual(MissingSourceValidation.ARGUMENT, "missing-source-validation")
        self.assertTrue(len(MissingSourceValidation.HELP) > 0)
        self.assertTrue(len(MissingSourceValidation.WIKI) > 0)

    def test_unrestricted_message_sender_metadata(self):
        self.assertEqual(UnrestrictedMessageSender.ARGUMENT, "unrestricted-message-sender")
        self.assertTrue(len(UnrestrictedMessageSender.HELP) > 0)
        self.assertTrue(len(UnrestrictedMessageSender.WIKI) > 0)

if __name__ == "__main__":
    unittest.main()
