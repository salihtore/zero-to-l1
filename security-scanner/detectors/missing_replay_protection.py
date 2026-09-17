from slither.detectors.abstract_detector import AbstractDetector, DetectorClassification
from slither.core.solidity_types.mapping_type import MappingType

class MissingReplayProtection(AbstractDetector):
    ARGUMENT = "missing-replay-protection"
    HELP = "Detects missing replay protection in Teleporter cross-chain message processing functions"
    IMPACT = DetectorClassification.HIGH
    CONFIDENCE = DetectorClassification.HIGH

    WIKI = "https://github.com/ava-labs/teleporter"
    WIKI_TITLE = "Missing Replay Protection in Teleporter Receiver"
    WIKI_DESCRIPTION = "Teleporter message receiver function does not check or update a state variable to track executed message IDs."
    WIKI_EXPLOIT_SCENARIO = "An attacker replays the same valid cross-chain message multiple times to trigger duplicate execution and drain funds."
    WIKI_RECOMMENDATION = "Maintain a state mapping/set of executed message IDs and revert if a message ID has already been executed."

    def _detect(self):
        results = []
        for contract in self.slither.contracts:
            for function in contract.functions:
                if function.is_constructor or function.is_fallback or function.is_receive:
                    continue
                
                is_teleporter_fn = (
                    "receiveTeleporterMessage" in function.name or
                    "teleporter" in function.name.lower() or
                    "teleporter" in contract.name.lower()
                )

                if is_teleporter_fn:
                    state_writes = function.state_variables_written
                    has_mapping_write = any(isinstance(var.type, MappingType) for var in state_writes)
                    
                    mapping_checked_or_written = has_mapping_write or any(
                        any(isinstance(v.type, MappingType) for v in node.state_variables_written + node.state_variables_read)
                        for node in function.nodes
                    )

                    if not mapping_checked_or_written:
                        info = [
                            function,
                            f" in contract {contract.name} lacks replay protection state tracking for cross-chain message IDs.\n",
                            "Neden Tehlikeli: İşlenen mesaj ID'leri durum değişkeninde (mapping/set) takip edilmediği için saldırganlar aynı cross-chain mesajını tekrar tekrar çalıştırarak kontrat bakiyesini boşaltabilir veya yetkisiz işlemleri yineleyebilir.\n"
                        ]
                        res = self.generate_result(info)
                        results.append(res)
        return results

