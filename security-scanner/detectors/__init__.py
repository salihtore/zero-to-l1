from .missing_replay_protection import MissingReplayProtection
from .missing_source_validation import MissingSourceValidation
from .unrestricted_message_sender import UnrestrictedMessageSender

CUSTOM_DETECTORS = [
    MissingReplayProtection,
    MissingSourceValidation,
    UnrestrictedMessageSender,
]
