# Security Scanner Module

Bu modül, Avalanche Interchain Messaging (ICM / Teleporter) entegrasyonlarına özel akıllı sözleşme güvenlik taramalarını gerçekleştirir.

## Özellikler & Yapı

- **Slither Tabanlı Taramalar**: Slither static analysis altyapısı üzerine inşa edilen özel (custom) dedektörler içerir.
- **ICM/Teleporter Dedektörleri**: `detectors/` klasörü altında Teleporter mesaj doğrulama, yetkilendirme ve cross-chain güvenliği kontrollerini yapan Python sınıfları yer alacaktır.
- **CLI Entegrasyonu**: Go CLI tarafından bir `subprocess` olarak çağrılarak tarama sonuçlarını JSON formatında orkestratöre iletir.

