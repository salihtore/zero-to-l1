// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/**
 * @title BadTeleporterReceiver
 * @notice Vulnerable contract for testing Teleporter security detectors.
 * Intentionally contains 3 major vulnerabilities:
 * 1. UnrestrictedMessageSender: external without access control modifier
 * 2. MissingSourceValidation: no check on originChainID or originSender
 * 3. MissingReplayProtection: no state tracking of executed message IDs
 */
contract BadTeleporterReceiver {
    event MessageProcessed(bytes32 indexed messageID, string message);

    // VULNERABLE:
    // - No modifier like `onlyTeleporter` (UnrestrictedMessageSender)
    // - No check for `originChainID` or `originSender` (MissingSourceValidation)
    // - No mapping to track `messageID` execution (MissingReplayProtection)
    function receiveTeleporterMessage(
        bytes32 originChainID,
        address originSender,
        bytes32 messageID,
        bytes calldata messageData
    ) external {
        string memory message = string(messageData);
        emit MessageProcessed(messageID, message);
    }
}

