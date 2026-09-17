from slither.detectors.abstract_detector import AbstractDetector, DetectorClassification

class MissingSourceValidation(AbstractDetector):
    ARGUMENT = "missing-source-validation"
    HELP = "Detects missing origin chain ID or origin sender validation in Teleporter message receivers"
    IMPACT = DetectorClassification.HIGH
    CONFIDENCE = DetectorClassification.HIGH

    WIKI = "https://github.com/ava-labs/teleporter"
    WIKI_TITLE = "Missing Origin Source Validation"
    WIKI_DESCRIPTION = "Teleporter message receiver function accepts cross-chain messages without verifying originChainID or originSender against an allowlist/require check."
    WIKI_EXPLOIT_SCENARIO = "An attacker sends cross-chain messages from an untrusted chain or malicious contract to trigger unauthorized execution."
    WIKI_RECOMMENDATION = "Add explicit require checks validating originChainID and originSender against trusted addresses."

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
                    param_names = [p.name.lower() for p in function.parameters]
                    has_origin_params = any("origin" in p or "chain" in p or "sender" in p for p in param_names)

                    if has_origin_params:
                        has_validation_check = False
                        for node in function.nodes:
                            node_str = str(node).lower()
                            if ("require" in node_str or "if" in node_str) and any(p in node_str for p in ["origin", "chain", "sender"]):
                                if "==" in node_str or "!=" in node_str:
                                    has_validation_check = True
                                    break

                        if not has_validation_check:
                            info = [
                                function,
                                f" in contract {contract.name} does not validate originChainID or originSender source address.\n",
                                "Neden Tehlikeli: Kaynak ağ (originChainID) ve gönderici adresi (originSender) doğrulanmadığı için yetkisiz veya kötü niyetli herhangi bir zincir/kontrat tarafından gönderilen mesajlar koşulsuz kabul edilir.\n"
                            ]
                            res = self.generate_result(info)
                            results.append(res)
        return results

