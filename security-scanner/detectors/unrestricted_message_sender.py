from slither.detectors.abstract_detector import AbstractDetector, DetectorClassification

class UnrestrictedMessageSender(AbstractDetector):
    ARGUMENT = "unrestricted-message-sender"
    HELP = "Detects Teleporter message receiver functions exposed without access control modifiers"
    IMPACT = DetectorClassification.HIGH
    CONFIDENCE = DetectorClassification.HIGH

    WIKI = "https://github.com/ava-labs/teleporter"
    WIKI_TITLE = "Unrestricted Message Sender Access Control"
    WIKI_DESCRIPTION = "Teleporter message receiver function is public/external without access control modifier (e.g. onlyTeleporter) or msg.sender check."
    WIKI_EXPLOIT_SCENARIO = "An attacker calls receiveTeleporterMessage directly on the local chain, bypassing the Teleporter protocol."
    WIKI_RECOMMENDATION = "Restrict function call access to the official Teleporter Messenger contract via modifier or msg.sender check."

    def _detect(self):
        results = []
        for contract in self.slither.contracts:
            for function in contract.functions:
                if function.is_constructor or function.is_fallback or function.is_receive:
                    continue

                if function.visibility in ["public", "external"]:
                    is_teleporter_fn = (
                        "receiveTeleporterMessage" in function.name or
                        "teleporter" in function.name.lower()
                    )

                    if is_teleporter_fn:
                        modifiers = [m.name.lower() for m in function.modifiers]
                        has_access_modifier = any("only" in m or "auth" in m or "teleporter" in m or "owner" in m for m in modifiers)

                        has_msg_sender_check = False
                        for node in function.nodes:
                            node_str = str(node).lower()
                            if "msg.sender" in node_str and ("==" in node_str or "!=" in node_str):
                                has_msg_sender_check = True
                                break

                        if not (has_access_modifier or has_msg_sender_check):
                            info = [
                                function,
                                f" in contract {contract.name} is public/external without access control modifier (onlyTeleporter/onlyOwner).\n",
                                "Neden Tehlikeli: Fonksiyonda erişim kısıtlaması modifier'ı (onlyTeleporter/onlyOwner) bulunmadığından yerel ağdaki herhangi bir kullanıcı Teleporter protokolünü bypass ederek bu fonksiyonu doğrudan çağırabilir.\n"
                            ]
                            res = self.generate_result(info)
                            results.append(res)
        return results

