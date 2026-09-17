# Zero to Secure L1

Avalanche üzerinde tek komutla L1/subnet kurulumu yapan, kurulum sırasında Interchain Messaging (ICM/Teleporter) güvenlik taraması gerçekleştiren ve kurulum sonrası canlı izleme paneli sunan uçtan uca platform.

## Faz Durumu (Checklist)

- [ ] **Faz 1: Launcher** (CLI ile subnet/L1 kurulumu, Avalanche-CLI üzerine orkestrasyon katmanı)
- [ ] **Faz 2: Security Scanner** (Slither tabanlı, ICM/Teleporter'a özgü custom detector'lar)
- [ ] **Faz 3: Dashboard** (Canlı validator/mesaj izleme paneli)

## Yerel Geliştirme Ortamı Kurulumu

Projenin geliştirilebilmesi için aşağıdaki gereksinimler gereklidir:

- **Go**: `v1.21+`
- **Python**: `v3.10+` (Slither tarayıcısı için)
- **Node.js**: `v18+` / **npm**: `v9+` (Dashboard frontend için)

### 1. CLI Modülü
```bash
cd cli
go mod download
go run cmd/main.go launch
```

### 2. Security Scanner Modülü
```bash
cd security-scanner
python -m venv venv
# Windows: venv\Scripts\activate | Linux/macOS: source venv/bin/activate
pip install -r requirements.txt
```

### 3. Dashboard Modülü

**Backend:**
```bash
cd dashboard/backend
go run main.go
```

**Frontend:**
```bash
cd dashboard/frontend
npm install
npm run dev
```

## Proje Dizin Yapısı

```
zero-to-l1/
├── cli/                 # Go + Cobra tabanlı CLI & orkestrasyon katmanı
├── security-scanner/    # Python + Slither tabanlı güvenlik tarayıcısı
├── dashboard/           # Izleme paneli backend ve frontend
├── docs/                # Dokümantasyon (litepaper ve mimari)
├── .gitignore           # Git ignore kuralları
├── README.md            # Proje özeti ve rehber
└── LICENSE              # MIT Lisansı
```

