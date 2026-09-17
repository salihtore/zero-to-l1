# Zero to Secure L1 - Mimari Özet

## Genel Bakış

Zero to Secure L1, Avalanche ağında tek bir CLI komutu ile L1/subnet kurulumu gerçekleştiren, kurulum esnasında interchain messaging (ICM/Teleporter) güvenlik taramaları yapan ve kurulum sonrasında canlı doğrulayıcı (validator) ile mesaj izleme paneli sunan bütünleşik bir platformdur.

## Hedef Ağ

- **Avalanche Fuji Testnet** (Mainnet hedef dışıdır)

## Teknoloji Kararları

1. **CLI + Orkestrasyon Katmanı (Faz 1)**:
   - **Dil / Kütüphane**: Go (`github.com/spf13/cobra`)
   - **Görevi**: Kullanıcı etkileşimi, Avalanche-CLI komutlarının orkestrasyonu, ağ kurulumu ve modüller arası koordinasyon.

2. **Security Scanner (Faz 2)**:
   - **Dil / Araç**: Python (`slither-analyzer` üzerine özel dedektörler)
   - **Görevi**: Interchain Messaging (ICM / Teleporter) entegrasyonlarına özel akıllı sözleşme güvenlik zafiyetlerini tespit etmek. Go CLI tarafından `subprocess` olarak tetiklenir.

3. **Dashboard Backend (Faz 3)**:
   - **Dil**: Go (veya Node.js — Faz 3 aşamasında netleştirilecektir)
   - **Görevi**: Validator durumlarını, L1 metriklerini ve ICM mesaj trafiğini canlı izleme servisleri sunmak.

4. **Dashboard Frontend (Faz 3)**:
   - **Teknoloji**: React + TypeScript (Vite bundler)
   - **Görevi**: Kullanıcıya canlı izleme ve metrik görselleştirme paneli sunmak.

## Geliştirme Fazları

- **Faz 1: Launcher**: CLI ile subnet/L1 kurulumu (Avalanche-CLI üzerine orkestrasyon katmanı).
- **Faz 2: Security Scanner**: Slither tabanlı, ICM/Teleporter'a özgü custom detector'lar.
- **Faz 3: Dashboard**: Canlı validator ve mesaj izleme paneli.

