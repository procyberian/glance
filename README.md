# glance-cli

Go ile yazılmış modüler bir ISO indirme, doğrulama ve yükleme aracıdır.

Bu sürümle birlikte araç artık:

- FTP dizinlerini özyineli olarak tarar
- HTTP/HTTPS dizinlerini üst ve alt dizinlerle birlikte tarar
- ISO listesini checksum bilgisiyle numaralı gösterir
- Tekli, çoklu veya `all` seçimini destekler
- Yarıda kalan indirmeleri `.download` dosyasından devam ettirir
- `--no-resume` ile sıfırdan yeniden başlatır
- `--output-path` ile tek ISO için tam hedef dosya yolunu destekler
- Etkileşimli kullanımda ISO seçiminden sonra hedef dizini sorar

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
- `--iso`, `--url`: HTTP, HTTPS, FTP veya yerel kaynak kabul eder
- `--checksum`: Verilen checksum ile doğrular; verilmezse uygun durumda otomatik bulur veya yerel olarak hesaplar
- `--checksum-algo`: `sha256`, `sha512`, `md5`
- Canlı ilerleme: yüzde, anlık hız, ortalama hız, ağ hızı, ETA
- FTP tarama: toplam ISO sayısını önce bulur, sonra checksum çözümü için tek ilerleme çubuğu gösterir
- HTTP/HTTPS tarama: genel erişime açık dizin sayfalarını özyineli gezer, aynı sunucu üzerinde üst ve alt dizinleri dikkate alır
- Seçim: `1`, `2,4,7`, `all`
- Sürdürme: `.download` uzantılı geçici dosya ile kaldığı yerden devam eder
- `--no-resume`: mevcut `.download` dosyasını yok sayıp sıfırdan indirir
- `--output`: hedef klasörü belirler
- `--output-path`: tek ISO için tam hedef dosya yolunu belirler
- Etkileşimli mod: seçim yapıldıktan sonra hedef dizin sorusu sorar
- `--upload`: SSH/SFTP ile dosya yükler
- Uzak checksum doğrulaması
- `--keygen`: `ed25519`, `rsa`, `ecdsa` anahtar üretir
- `--license`: MIT lisans metnini yazdırır

## Derleme

```bash
go mod tidy
go build -o glance .
```

## Son Kullanıcı Kullanım Rehberi

### 1. Doğrudan ISO indir

```bash
./glance --download --iso https://releases.ubuntu.com/24.04/ubuntu-24.04.4-live-server-amd64.iso
```

### 2. FTP dizinini tara, listele, seç ve indir

```bash
./glance --download --iso ftp://ftp.example.com/iso/
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

### 4. Seçim biçimleri

```text
1
1,3,5
all
```

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

- `dosya.iso.sha256sum` / `dosya.iso.sha512sum` / `dosya.iso.md5sum`
- `SHA256SUMS` / `SHA512SUMS` / `MD5SUMS`
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
- `--checksum`
- `--checksum-algo`
- `--output`
- `--output-path`
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
