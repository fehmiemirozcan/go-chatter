<div align="center">

# 🕵️‍♂️ Go-Chatter

### Uçtan Uca Şifreli, Sıfır İz Bırakan P2P Terminal Sohbet Uygulaması

[![Go Sürümü](https://img.shields.io/badge/Go-1.25%2B-00ADD8?style=flat&logo=go)](https://golang.org)
[![Şifreleme](https://img.shields.io/badge/%C5%9Eifreleme-AES--256--GCM-success?style=flat)](https://en.wikipedia.org/wiki/Galois/Counter_Mode)
[![Kayıt](https://img.shields.io/badge/Kay%C4%B1t-S%C4%B1f%C4%B1r%20Veri%20(Sadece%20RAM)-red?style=flat)](#-gizlilik-ve-sf-veri-politikas-zero-persistence)
[![TUI](https://img.shields.io/badge/TUI-Charmbracelet%20Bubble%20Tea-f43f5e?style=flat)](https://github.com/charmbracelet/bubbletea)
[![Mimari](https://img.shields.io/badge/A%C4%9F-P2P%20%2F%20Host%20Migration-blueviolet?style=flat)](#-a-mimarisi-ve-al-prensibi)

[ **English** ](README.md) | [ **Türkçe** ](README.tr.md)

</div>

---

## 🔒 Gizlilik ve Sıfır Veri Politikası (Zero-Persistence)

**Go-Chatter**, terminal üzerinden tamamen anonim, güvenli ve iz bırakmayan bir mesajlaşma deneyimi sunmak amacıyla geliştirilmiştir:

* **Veritabanı Yok:** SQL, NoSQL veya herhangi bir harici veri depolama birimi kullanılmaz.
* **Disk Günlüğü (Log) Yok:** Mesajlar, kullanıcı isimleri ve oda parolaları asla diske yazılmaz, günlük dosyalarında tutulmaz.
* **Sadece Bellek (RAM-Only):** Tüm iletişim doğrudan bellek kanalları (Go channels) ve anlık ağ tamponları (network buffers) üzerinden akar.
* **Oda Değişiminde Anında Temizlik:** Bir odadan başka bir odaya geçtiğiniz anda eski odaya ait mesaj geçmişi hafızadan tamamen silinir.
* **Çıkışta Sıfır İz:** Uygulama veya bağlantı kapatıldığı anda oturuma ait tüm veriler geride hiçbir iz bırakmadan yok olur.

---

## ✨ Öne Çıkan Özellikler

- **🌐 Global ve Özel Odalar:** Uygulamayı açar açmaz genel `global` odaya bağlanabilir veya `/createroom` ile şifreli özel odalar oluşturabilirsiniz.
- **🔐 AES-256-GCM Uçtan Uca Şifreleme:**
  - Parolalı odalarda gönderilen tüm mesajlar istemci tarafında **AES-256-GCM** algoritması ile kilitlenir.
  - Parolalardan 32-byte anahtar üretimi **SHA-256** ile sağlanır.
  - Her mesaj için `crypto/rand` ile 12-byte rastgele nonce ve 16-byte MAC doğrulama etiketi oluşturulur; böylece paket üzerinde oynanamaz veya aynı mesaj tekrar iletilemez.
- **🔄 P2P Otomatik Keşif ve Dinamik Lider Seçimi:**
  - Sabit veya merkezi bir sunucu kurulumu gerektirmez.
  - Ağı ilk başlatan istemci **UDP yayını** aracılığıyla lider arar; lider yoksa otomatik olarak kendisi **Lider** olur ve TCP sunucusunu açar.
  - Sonraki kullanıcılar UDP keşfiyle lideri otomatik bulur ve bağlanır.
- **⚡ Hızlı Host Migration (Lider Değişimi):**
  - Lider sohbetten çıktığında veya bağlantısı koptuğunda, bağlı olan diğer istemciler anında yeni bir lider seçimi başlatır ve sunucu devralınarak sohbet kesintisiz devam eder.
- **💬 Fısıltı (Özel 1'e 1 Mesajlaşma):**
  - Hangi odada olursanız olun `/whisper <kullanici_adi> : <mesaj>` komutu ile ağdaki belirli bir kullanıcıya doğrudan gizli mesaj iletebilirsiniz.
- **🟢 Çevrimiçi Durum Yönetimi:**
  - `/status` komutuyla durumunuzu anlık olarak değiştirebilirsiniz (`online`, `away`, `dnd`). Durumunuz sağ paneldeki kullanıcı listesinde anında güncellenir.
- **🛡️ Spam ve Kötüye Kullanım Korumaları:**
  - **Hız Sınırı (Rate Limit):** Hem istemci hem sunucu tarafında 300ms mesaj aralığı sınırı ile spam engellenir.
  - **IP Bağlantı Sınırı:** Ağ kaynaklarını korumak amacıyla aynı IP adresinden en fazla 3 bağlantıya izin verilir.
  - **Kullanıcı Adı Çakışma Koruması:** Harf büyüklüğü ve Türkçe karakter normalizasyonu ile aynı kullanıcı adının tekrar alınması engellenir.
- **🎨 Modern Terminal Arayüzü (TUI):**
  - **Charmbracelet Bubble Tea**, **Lipgloss** ve **Bubbles** kütüphaneleriyle hazırlanmıştır.
  - 3 bölmeli düzen: Mesajlaşma alanı, Odalar listesi ve Kullanıcılar listesi.
  - Tam **Fare (Mouse)** ve **Klavye** desteği.
  - Komut otomatik tamamlama ve geçmiş mesajlar arasında gezinti.

---

## 🏗️ Ağ Mimarisi ve Çalışma Prensibi

```
                      +-----------------------------+
                      |    Uygulama Başlatılır      |
                      +-----------------------------+
                                     |
                       [:8081 Portunda UDP Yayını]
                       "who_is_the_leader" sorgusu
                                     |
                   +-----------------+-----------------+
                   |                                   |
           (Yanıt Yok / Zaman Aşımı)           ("I_am_the_leader" Yanıtı)
                   |                                   |
                   v                                   v
        +----------------------+            +----------------------+
        |      LİDER OL        |            |     TAKİPÇİ OL       |
        | - UDP :8081 dinle    |            | - Liderin :9000      |
        | - TCP :9000 başlat   |            |   TCP portuna bağlan |
        | - Router Hub yönet   |            | - "global" odaya gir |
        +----------------------+            +----------------------+
                   ^                                   |
                   |      [Lider Bağlantısı Koptu]     |
                   +-----------------------------------+
                         Host Migration Başlatılır
```

### Port Detayları:
- **UDP Keşif Portu:** `8081` (Lider sorgulama ve cevaplama)
- **TCP Router Portu:** `9000` (Mesaj dağıtımı, oda yönetimi, fısıltı yönlendirmesi)

---

## ⌨️ Komut Listesi

| Komut | Kullanım Biçimi | Açıklama |
| :--- | :--- | :--- |
| `/help` | `/help` | Kullanılabilir komutları ve yardım mesajını listeler. |
| `/createroom` | `/createroom <oda_adi> [şifre]` | Yeni bir oda oluşturur (kullanıcı başına en fazla 3 özel oda). Şifre verilirse AES-256-GCM ile şifrelenir. |
| `/join` | `/join <oda_adi> [şifre]` | Mevcut bir odaya geçiş yapar. Şifreli odalarda şifre doğrulaması yapılır. |
| `/whisper` | `/whisper <kullanici_adi> : <mesaj>` | Belirli bir kullanıcıya doğrudan özel fısıltı mesajı gönderir. |
| `/status` | `/status <online\|away\|dnd>` | Kullanıcı durumunu günceller (`online`, `away`, `dnd`). |
| `/quit` | `/quit` | Bağlantıyı güvenli bir şekilde kapatır ve uygulamadan çıkar. |

---

## 🎮 Arayüz Kontrolleri ve Kısayollar

| Tuş / Fare Hareketi | İşlev |
| :--- | :--- |
| `Tab` | **Mesaj Girişi** ➔ **Sohbet Alanı** ➔ **Odalar Paneli** ➔ **Kullanıcılar Paneli** arasında odak değiştirir.<br>*(Girdi `/` ile başlıyorsa komutları otomatik tamamlar).* |
| `Esc` | Aktif arama filtresini sıfırlar veya odağı anında mesaj giriş çubuğuna geri getirir. |
| `Yukarı` / `Aşağı` | Mesaj çubuğundayken: Daha önce gönderilen mesajlar arasında gezinir.<br>Listelerde/Sohbette: Satırları kaydırır. |
| `Enter` | Mesaj gönderir, seçili odaya katılır veya seçili kullanıcıya fısıltı başlatır. |
| `/` (Odalar/Kullanıcılar paneli) | Odaklanılan listede anında interaktif filtreleme / arama başlatır. |
| `Fare Sol Tık` | Herhangi bir panele odaklanır veya oda/kullanıcı seçer. Şifreli odaya tıklandığında `/join` komutu, kullanıcıya tıklandığında `/whisper` komutu hazırlanır. |
| `Fare Tekerleği` | Sohbet akışını veya oda/kullanıcı listelerini yukarı-aşağı kaydırır. |
| `Ctrl + C` | Uygulamayı zorla kapatır. |

---

## 🛠️ Teknoloji Yığını ve Bağımlılıklar

Proje yalnızca modern Go standart kütüphaneleri ve Charmbracelet TUI ekosistemi kullanılarak geliştirilmiştir:

* **Temel Dil:** [Go](https://go.dev/) (`1.25+` / standart kütüphane)
  - `net`: TCP router ve UDP discovery soket yönetimi.
  - `crypto/aes`, `crypto/cipher`, `crypto/sha256`, `crypto/rand`: AES-256-GCM şifreleme ve güvenli anahtar yönetimi.
  - `sync`: Eşzamanlılık (concurrency) ve veri güvenliği için Mutex kilitleri.
* **Terminal Arayüzü (TUI):**
  - `github.com/charmbracelet/bubbletea`: TUI yaşam döngüsü ve olay yönetimi.
  - `github.com/charmbracelet/lipgloss`: Stil tanımları, kenarlıklar ve renk motoru.
  - `github.com/charmbracelet/bubbles`: Giriş çubuğu (textinput), kaydırma alanı (viewport) ve listeler (list).

---

## 🚀 Kurulum ve Çalıştırma

### Gereksinimler
* Bilgisayarınızda yüklü [Go](https://go.dev/dl/) (`1.21` veya üzeri tavsiye edilir, `1.25.1` ile test edilmiştir).

### Çalıştırma Adımları

1. **Depoyu klonlayın:**
   ```bash
   git clone https://github.com/fehmiemirozcan/go-chatter.git
   cd go-chatter
   ```

2. **Uygulamayı çalıştırın:**
   ```bash
   go run ./cmd/main.go
   ```
   *Veya Makefile kullanarak:*
   ```bash
   make run
   ```

3. **Birden Fazla Kullanıcıyla Deneyin:**
   - Aynı bilgisayarda veya aynı yerel ağdaki başka bir terminal açın:
     ```bash
     go run ./cmd/main.go
     ```
   - Farklı bir kullanıcı adı belirleyin ve hemen sohbete başlayın.

---

## 🧪 Testler

Uygulamanın istemci, sunucu, komutlar, güvenlik ve arayüz modüllerini test etmek için:

```bash
# Tüm birim ve entegrasyon testlerini çalıştır
make test
# Veya doğrudan Go komutuyla:
go test ./...
```

Test kapsamını (coverage) görüntülemek için:
```bash
make coverage
```

---

## 📁 Dizin Yapısı

```
go-chatter/
├── cmd/
│   └── main.go                  # Giriş noktası, lider kontrolü ve başlatma
├── internal/
│   ├── client/
│   │   ├── chat.go              # İstemci mesaj döngüsü, ayrıştırıcı ve migration mantığı
│   │   └── chat_test.go         # İstemci birim testleri
│   ├── command/
│   │   ├── command_exit.go      # /quit komutu
│   │   ├── command_get.go       # Komut sözlüğü ve handler tanımları
│   │   ├── command_help.go      # /help yardım komutu
│   │   ├── command_roomactual.go# /createroom ve /join işlemleri
│   │   ├── command_status.go    # /status durum komutu
│   │   ├── command_whisper.go   # /whisper doğrudan mesajlaşma
│   │   └── *_test.go            # Komut testleri
│   ├── discovery/
│   │   ├── discovery.go         # UDP yayın ve lider keşfi (:8081)
│   │   └── discovery_test.go    # Keşif testleri
│   ├── security/
│   │   ├── security.go          # SHA-256 anahtar türetme, AES-256-GCM Şifreleme/Çözme
│   │   └── security_test.go     # Kriptografi testleri
│   ├── server/
│   │   ├── server.go            # RouterHub, TCP dinleyici (:9000), oda ve kullanıcı yönetimi
│   │   └── server_test.go       # Eşzamanlılık ve protokol testleri
│   └── tui/
│       ├── styles.go            # Lipgloss stil ve renk tanımları
│       ├── tui.go               # Bubble Tea başlatma
│       ├── types.go             # TUI veri tipleri ve mesaj tanımları
│       ├── update.go            # Tuş/fare olayları, odak yönetimi
│       ├── view.go              # Panel çizimleri (Sohbet, Odalar, Kullanıcılar)
│       └── listenMessage_test.go# TUI mesaj ayrıştırma testleri
├── Makefile                     # Çalıştırma, test ve coverage kısayolları
├── go.mod                       # Modül ve bağımlılık tanımları
└── go.sum                       # Bağımlılık sağlama toplamları
```
