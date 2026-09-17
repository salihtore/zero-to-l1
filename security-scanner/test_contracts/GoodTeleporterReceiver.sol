// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/**
 * @title GoodTeleporterReceiver
 * @notice Secure contract implementing proper Teleporter security protections:
 * 1. Access control via `onlyTeleporter` modifier
 * 2. Origin source validation (allowedSourceChainID & allowedSender)
 * 3. Replay protection using state mapping `executedMessages`
 */
contract GoodTeleporterReceiver {
    address public immutable teleporterMessenger;
    bytes32 public immutable allowedSourceChainID;
    address public immutable allowedSender;

    mapping(bytes32 => bool) public executedMessages;

    event MessageProcessed(bytes32 indexed messageID, string message);

    constructor(address _teleporterMessenger, bytes32 _allowedSourceChainID, address _allowedSender) {
        teleporterMessenger = _teleporterMessenger;
        allowedSourceChainID = _allowedSourceChainID;
        allowedSender = _allowedSender;
    }

    modifier onlyTeleporter() {
        require(msg.sender == teleporterMessenger, "GoodTeleporterReceiver: Unauthorized sender");
        _;
    }

    function receiveTeleporterMessage(
        bytes32 originChainID,
        address originSender,
        bytes32 messageID,
        bytes calldata messageData
    ) external onlyTeleporter {
        // Source Validation
        require(originChainID == allowedSourceChainID, "GoodTeleporterReceiver: Invalid origin chain");
        require(originSender == allowedSender, "GoodTeleporterReceiver: Invalid origin sender");

        // Replay Protection
        require(!executedMessages[messageID], "GoodTeleporterReceiver: Message already executed");
        executedMessages[messageID] = true;

        string memory message = string(messageData);
        emit MessageProcessed(messageID, message);
    }
}

