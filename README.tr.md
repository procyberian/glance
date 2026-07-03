# glance-cli

Go ile yazılmış modüler bir ISO indirme, doğrulama ve yükleme aracıdır.

Bu proje, yapay zeka destekli Microsoft Copilot yardımıyla başlatılmıştır; ancak insan yönlendirmesiyle sürekli olarak güncellenmektedir. Proje, MIT lisansı altında Özgür Yazılım olarak sunulur ve katkılara açıktır. PlusClouds, topluluktan gelen katkıları arkadaşça memnuniyetle kabul eder.

## Yayınlanan Sürüm

`glance`, artık `v11` Go modül hattı ve `v11.0.8` git sürüm etiketi ile yayımlanmıştır.

En güncel `v11` sürümünü doğrudan şu komutla kurabilirsiniz:

```bash
go install github.com/procyberian/glance/v11@latest
```

Bu komut release arşivlerinden birini indirmez. Mevcut makinenin `GOOS` ve `GOARCH` değerlerine göre yerelde derleme yapar; yani oluşan binary kullanıcının kendi mimarisine uygun olur.

Kaynak kodu indirip yerelde derlemek için:

```bash
git clone git@github.com:procyberian/glance.git
cd glance
git checkout v11.0.8
go build -o glance .
```

## Binary İndirmeleri

`v11.0.8` sürümünün derlenmiş binary arşivleri proje yayın sayfalarında dağıtılır:

- GitHub Releases: <https://github.com/procyberian/glance/releases>
- Codeberg Releases: <https://codeberg.org/procyberian/glance/releases>

`v11.0.8` sürümünü açıp platformunuza uygun asset dosyasını indirin.

Mimari eşlemesi:

- Linux x86_64: `glance-linux-amd64.tar.gz`
- Linux ARM64: `glance-linux-arm64.tar.gz`
- macOS Intel: `glance-darwin-amd64.tar.gz`
- macOS Apple Silicon: `glance-darwin-arm64.tar.gz`
- Windows x86_64: `glance-windows-amd64.zip`

Bu sürüm için planlanan asset adları:

- `glance-linux-amd64.tar.gz`
- `glance-linux-arm64.tar.gz`
- `glance-darwin-amd64.tar.gz`
- `glance-darwin-arm64.tar.gz`
- `glance-windows-amd64.zip`

API token'ları tanımlandıktan sonra release kaydını oluşturup asset dosyalarını otomatik yüklemek için:

```bash
GH_TOKEN=... CODEBERG_TOKEN=... ./scripts/publish-release.sh v11.0.8
```

Token kapsamı önerileri:

- GitHub klasik personal access token: `repo`
- GitHub fine-grained token: repository `Contents` izni read/write
- Codeberg token: release ve release asset işlemleri için repository write erişimi

Yararlı script modları:

```bash
./scripts/publish-release.sh --dry-run v11.0.8
GH_TOKEN=... ./scripts/publish-release.sh --github-only v11.0.8
CODEBERG_TOKEN=... ./scripts/publish-release.sh --codeberg-only v11.0.8
GH_TOKEN=... CODEBERG_TOKEN=... ./scripts/publish-release.sh --dist-dir dist --notes-file release-notes/v11.0.8.md v11.0.8
```

Bu sürümle birlikte:

- `scripts/publish-release.sh` betiğine etkileşimli token sorgulama eklendi: `GH_TOKEN` veya `CODEBERG_TOKEN` ortam değişkeni tanımlı değilse betik hemen hata verip çıkmak yerine terminalde gizli girişli (`read -rsp`) bir parola istemi gösterir

## Kimler İçin?

### Son Kullanıcılar

- Hazır ISO listesinden seçim yaparak manuel URL kopyalama ihtiyacını azaltır
- Checksum doğrulamasını otomatik yapar
- Yarıda kalan indirmelerde tekrar baştan başlama zorunluluğunu ortadan kaldırır

### Sistem Yöneticileri

- Ayna sunucu veya kurumsal depo üzerinden ISO seçimini standartlaştırır
- SSH/SFTP ile hedef makineye aktarım yapar
- `known_hosts` doğrulaması ve uzak checksum karşılaştırması sunar

### Geliştiriciler

- Go modüllü yapı ile CLI, downloader, verifier, uploader ve keygen katmanlarına ayrılmıştır
- HTTP ve FTP tarama davranışları ayrı fonksiyonlarda korunur
- Değişiklik günlüğü ve README dosyaları sürüm hazırlama sürecini destekler

## Özellikler

- `--download`: ISO indirir veya yerel ISO dosyasını kopyalar
- `--iso`, `--url`: HTTP, HTTPS, FTP veya yerel kaynak kabul eder (`ftp://` için `--allow-insecure-ftp` gerekir)
- `--checksum`: Verilen checksum ile doğrular; verilmezse uygun durumda otomatik bulur veya yerel olarak hesaplar
- `--checksum-algo`: `sha256`, `sha512`
- Canlı ilerleme: yüzde, anlık hız, ortalama hız, ağ hızı, ETA
- FTP tarama: toplam ISO sayısını önce bulur, sonra checksum çözümü için tek ilerleme çubuğu gösterir
- HTTP/HTTPS tarama: genel erişime açık dizin sayfalarını özyineli gezer, aynı sunucu üzerinde üst ve alt dizinleri dikkate alır
- Seçim: `1`, `2,4,7`, `all`
- Sürdürme: `.download` uzantılı geçici dosya ile kaldığı yerden devam eder
- `--no-resume`: mevcut `.download` dosyasını yok sayıp sıfırdan indirir
- `--output`: hedef klasörü belirler
- `--output-path`: tek ISO için tam hedef dosya yolunu belirler
- `--scan-timeout`: HTTP/FTP dizin taraması için zaman aşımı (saniye)
- Etkileşimli mod: seçim yapıldıktan sonra hedef dizin sorusu sorar
- Etkileşimli upload: `--file` verilmezse yerel ISO/imaj dosyalarını numaralı listeler, seçim ister
- `--upload`: SSH/SFTP ile dosya yükler
- Uzak checksum doğrulaması
- `--keygen`: `ed25519`, `rsa`, `ecdsa` anahtar üretir
- `--license`: MIT lisans metnini yazdırır

## Derleme

```bash
go mod tidy
go build -o glance .
```

## Kütüphane Kullanımı

Etiketlenmiş proje artık `github.com/procyberian/glance/v11/pkg/glance` yolunda tekrar kullanılabilir bir genel paket sunar.

Örnek:

```go
package main

import (
    "log"

    glance "github.com/procyberian/glance/v11/pkg/glance"
)

func main() {
    result, err := glance.DownloadAndVerify(glance.DownloadOptions{
        Source:            "https://releases.ubuntu.com/24.04/ubuntu-24.04.4-live-server-amd64.iso",
        OutputDir:         "./downloads",
        ChecksumAlgorithm: "sha256",
        AllowResume:       true,
    })
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("downloaded=%s checksum=%s algo=%s", result.Path, result.Checksum, result.Algorithm)
}
```

Bu paket ayrıca `DownloadISO`, `ListFTPISOs`, `ListHTTPISOs`, `ResolveChecksum`, `VerifyFileHash`, `CalculateFileHash`, `UploadFile`, `GenerateKeyPair`, `Parse`, `Run` ve `Execute` sarmalayıcılarını da dışa açar.

## Son Kullanıcı Kullanım Rehberi

### 1. Doğrudan ISO indir

```bash
./glance --download --iso https://releases.ubuntu.com/24.04/ubuntu-24.04.4-live-server-amd64.iso
```

### 2. FTP dizinini tara, listele, seç ve indir

```bash
./glance --download --allow-insecure-ftp --iso ftp://ftp.example.com/iso/
```

Beklenen akış:

```text
Detecting total ISO count...
Found 100 total ISOs
[==========----------] 50.00% | 50/100 ISOs
FTP ISO list:
    1) example-1.iso | size: 2.80 GB | checksum: ...
    2) example-2.iso | size: 1.90 GB | checksum: ...
Select ISO number(s) (example: 1 or 1,3,5 or all) [1-100]:
Which directory should the ISO be downloaded to? [default: ./downloads]:
```

### 3. HTTP/HTTPS dizinini özyineli tara

```bash
./glance --download --iso https://ftp.example.com/
```

Bu modda araç:

- Genel erişime açık dizini okur
- Aynı sunucu üzerindeki ilgili alt dizinleri gezer
- Başlangıç yolunun üst dizinlerini de taramaya dahil eder
- Gerçek `.iso` dosyalarını seçim listesine ekler

### 3.1 Doğrudan FTP ISO dosyasını indir

```bash
./glance --download --allow-insecure-ftp --iso ftp://ftp.example.com/iso/example.iso
```

Bu kullanımda araç FTP sunucusundan dosyayı doğrudan indirir. Yarım kalan `.download` dosyası varsa sürdürmeyi dener; sunucu desteklemiyorsa güvenli şekilde sıfırdan yeniden başlatır.

### 4. Seçim biçimleri

```text
1
1,3,5
all
```

### 4.1 Dizin tarama zaman aşımı ayarla

```bash
./glance --download --iso https://ftp.uni-stuttgart.de/debian-cd/current/amd64/iso-cd/ --scan-timeout 180
```

Not:

- Varsayılan değer `60` saniyedir
- `--scan-timeout 0` verilirse zaman aşımı devre dışı kalır

Sorun giderme (argüman sırası):

- Yanlış kullanım: `./glance --download --iso --scan-timeout https://ftp.uni-stuttgart.de/debian-cd/`
- Doğru kullanım: `./glance --download --iso https://ftp.uni-stuttgart.de/debian-cd/ --scan-timeout 180`
- Hata durumunda araç artık şu tip bir mesaj verir: `unexpected positional arguments: ...`

### 5. Yarıda kalan indirmeye devam et

```bash
./glance --download --iso https://www.ututo.org/downloads/Candidato-Ututo-2017-UL.iso
```

Eğer `downloads/Candidato-Ututo-2017-UL.iso.download` varsa araç şu şekilde davranır:

```text
Resume file found: downloads/Candidato-Ututo-2017-UL.iso.download (44.69 MB). Download will continue from where it left off.
Starting ISO download (1/1)...
Resuming download from 44.69 MB: downloads/Candidato-Ututo-2017-UL.iso.download
```

### 6. Sürdürme özelliğini kapat ve sıfırdan başlat

```bash
./glance --download --no-resume --iso https://www.ututo.org/downloads/Candidato-Ututo-2017-UL.iso
```

### 7. ISO'yu belirli bir klasöre indir

```bash
./glance --download --iso https://www.ututo.org/downloads/Candidato-Ututo-2017-UL.iso --output /srv/iso-cache
```

### 8. ISO'yu tam istediğin dosya yoluna indir

```bash
./glance --download --iso https://www.ututo.org/downloads/Candidato-Ututo-2017-UL.iso --output-path /srv/releases/custom-ututo.iso
```

Not:

- `--output-path` sadece tek ISO için geçerlidir
- Çoklu seçimde `--output` kullanılmalıdır

### 9. Yerel ISO'yu kopyala ve checksum hesapla

```bash
./glance --download --iso /home/user/isos/archlinux-x86_64.iso
```

### 10. Belirli checksum ile doğrula

```bash
./glance --download --iso https://releases.ubuntu.com/24.04/ubuntu-24.04.4-live-server-amd64.iso --checksum e907d92eeec9df64163a7e454cbc8d7755e8ddc7ed42f99dbc80c40f1a138433
```

## Sistem Yöneticisi Rehberi

### Ayna sunucu ve depo kullanım önerileri

- FTP tarafında anonim giriş desteklenmeyen hostlarda HTTP/HTTPS taraması tercih edilmelidir
- Büyük ayna sunucularda tarama sırasında hız sınırlaması uygulanır
- FTP taramada dizin istekleri arasında gecikme kullanılır; bu sayede saldırgan trafik gibi görünme riski azaltılır

### SSH/SFTP yükleme

Şifre ile:

```bash
./glance --upload --file ./downloads/ubuntu-24.04.4-live-server-amd64.iso --host 192.168.1.50 --user root --password secret
```

SSH anahtarı ile:

```bash
./glance --upload --file ./downloads/ubuntu-24.04.4-live-server-amd64.iso --host 192.168.1.50 --user root --ssh-key ~/.ssh/id_rsa --known-hosts ~/.ssh/known_hosts
```

`--file` olmadan etkileşimli seçim:

```bash
./glance --upload --iso https://releases.ubuntu.com/24.04/
```

Beklenen akış:

```text
Local ISO/file candidates for upload:
    1) downloads/ubuntu-24.04.4-live-server-amd64.iso
    2) downloads/debian-12.11.0-amd64-netinst.iso
Select file number [1-2] or type a full path:
SSH host/IP:
SSH username:
Auth method (password/key):
```

Tek komutta indir ve yükle:

```bash
./glance --download --iso https://releases.ubuntu.com/24.04/ubuntu-24.04.4-live-server-amd64.iso --upload --host 192.168.1.50 --user root --ssh-key ~/.ssh/id_rsa --known-hosts ~/.ssh/known_hosts
```

`known_hosts` kaydı yoksa önce ekleyin:

```bash
ssh-keyscan -H 192.168.1.50 >> ~/.ssh/known_hosts
```

### Operasyonel notlar

- Sürdürme dosyaları nihai dosya adının sonuna `.download` eklenerek tutulur
- İndirme tamamlanınca geçici dosya nihai dosyaya çevrilir
- Sunucu `Range` desteklemiyorsa araç otomatik olarak sıfırdan yeniden indirir
- FTP sunucusu resume (`REST`) desteklemiyorsa araç otomatik olarak sıfırdan yeniden indirir

## Geliştirici Rehberi

### Paket yapısı

- `internal/cli`: seçenek ayrıştırma, istemler, seçim ve akışın orkestrasyonu
- `internal/downloader`: HTTP/FTP tarama, indirme, sürdürme ve checksum kaynak çözümü
- `internal/verifier`: yerel checksum hesaplama ve doğrulama
- `internal/uploader`: SSH/SFTP yükleme ve uzak checksum doğrulama
- `internal/keygen`: SSH anahtar üretimi

### Önerilen geliştirme komutları

```bash
go build -o glance .
./glance --help
./glance --download --iso https://www.ututo.org/downloads/
```

### Checksum çözme sırası

HTTP/HTTPS için:

- `dosya.iso.sha256sum` / `dosya.iso.sha512sum`
- `SHA256SUMS` / `SHA512SUMS`
- `checksum`

FTP için:

- aynı dizindeki benzer checksum dosyaları
- indeks dosyaları ve yaygın checksum dosya adları

Bulunamazsa yerel hash hesaplanır.

### Etkileşimli karar akışı

1. Kaynak bir dizinse ISO listesi çıkarılır
2. Kullanıcı bir veya daha fazla ISO seçer
3. `--output` ya da `--output-path` verilmediyse hedef dizin sorulur
4. `.download` dosyası varsa sürdürme bilgisi gösterilir
5. İndirme biterse checksum doğrulama yapılır
6. `--upload` varsa uzak sisteme aktarılır
7. Upload sırasında `--file` yoksa yerel aday dosyalar listelenir ve kullanıcı seçim yapar

## Tüm Seçenekler

- `--download`
- `--no-resume`
- `--upload`
- `--keygen`
- `--key-algo`
- `--key-output`
- `--key-name`
- `--iso`
- `--url`
- `--allow-insecure-ftp`
- `--checksum`
- `--checksum-algo`
- `--output`
- `--output-path`
- `--scan-timeout`
- `--file`
- `--host`
- `--port`
- `--user`
- `--password`
- `--ssh-key`
- `--known-hosts`
- `--remote-path`
- `--license`
- `--help`

## Değişiklik Günlüğü Notu

`ChangeLog.md` sürüm notları ve günlük değişim kaydı için tutulur. En güncel commit zamanını doğrulamak için:

```bash
git log --date=iso --pretty=format:"%h %ad %an: %s"
```

## License

MIT License

Copyright (c) 2026 PlusClouds

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
