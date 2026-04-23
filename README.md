# glance-cli

Go ile yazilmis moduler bir ISO indirme, dogrulama ve yukleme aracidir.

Bu surumle birlikte arac artik:

- FTP dizinlerini recursive tarar
- HTTP/HTTPS dizinlerini ust ve alt dizinlerle birlikte tarar
- ISO listesini checksum bilgisiyle numarali gosterir
- Tekli, coklu veya `all` secim destekler
- Yarida kalan indirmeleri `.download` dosyasindan devam ettirir
- `--no-resume` ile sifirdan yeniden baslatir
- `--output-path` ile tek ISO icin tam hedef dosya yolunu destekler
- Etkilesimli kullanimda ISO seciminden sonra hedef dizini sorar

## Kimler Icin?

### Son Kullanicilar

- Hazir ISO listesinden secim yaparak manuel URL kopyalama ihtiyacini azaltir
- Checksum dogrulamasini otomatik yapar
- Yarida kalan indirmelerde tekrar bastan baslama zorunlulugunu ortadan kaldirir

### Sistem Yoneticileri

- Mirror veya kurumsal depo uzerinden ISO secimini standardize eder
- SSH/SFTP ile hedef makineye aktarim yapar
- `known_hosts` dogrulamasi ve uzak checksum karsilastirmasi sunar

### Gelistiriciler

- Go modullu yapi ile CLI, downloader, verifier, uploader ve keygen katmanlarina ayrilmistir
- HTTP ve FTP tarama davranislari ayri fonksiyonlarda korunur
- Changelog ve README dosyalari release hazirlama surecini destekler

## Ozellikler

- `--download`: ISO indirir veya lokal ISO dosyasini kopyalar
- `--iso`, `--url`: HTTP, HTTPS, FTP veya lokal kaynak kabul eder
- `--checksum`: Verilen checksum ile dogrular; verilmezse uygun durumda otomatik bulur veya lokal hesaplar
- `--checksum-algo`: `sha256`, `sha512`, `md5`
- Canli ilerleme: yuzde, anlik hiz, ortalama hiz, ag hizi, ETA
- FTP tarama: toplam ISO sayisini once bulur, sonra checksum cozumu icin tek progress bar gosterir
- HTTP/HTTPS tarama: public dizin sayfalarini recursive gezer, ayni host icinde ust ve alt dizinleri dikkate alir
- Secim: `1`, `2,4,7`, `all`
- Resume: `.download` uzantili gecici dosya ile kalan yerden devam
- `--no-resume`: mevcut `.download` dosyasini yok sayip sifirdan indirir
- `--output`: hedef klasoru belirler
- `--output-path`: tek ISO icin tam hedef dosya yolunu belirler
- Etkilesimli mod: secim yapildiktan sonra hedef dizin sorusu
- `--upload`: SSH/SFTP ile dosya yukler
- Uzak checksum dogrulamasi
- `--keygen`: `ed25519`, `rsa`, `ecdsa` anahtar uretir
- `--license`: MIT lisans metnini yazdirir

## Derleme

```bash
go mod tidy
go build -o glance .
```

## Son Kullanici Kullanim Rehberi

### 1. Dogrudan ISO indir

```bash
./glance --download --iso https://releases.ubuntu.com/24.04/ubuntu-24.04.4-live-server-amd64.iso
```

### 2. FTP dizinini tara, listele, sec ve indir

```bash
./glance --download --iso ftp://ftp.example.com/iso/
```

Beklenen akis:

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

### 3. HTTP/HTTPS dizinini recursive tara

```bash
./glance --download --iso https://ftp.example.com/
```

Bu modda arac:

- Public dizini okur
- Ayni host icindeki ilgili alt dizinleri gezer
- Baslangic yolunun ust dizinlerini de taramaya dahil eder
- Gercek `.iso` dosyalarini secim listesine ekler

### 4. Secim formatlari

```text
1
1,3,5
all
```

### 5. Yarida kalan indirmeye devam et

```bash
./glance --download --iso https://www.ututo.org/downloads/Candidato-Ututo-2017-UL.iso
```

Eger `downloads/Candidato-Ututo-2017-UL.iso.download` varsa arac su sekilde davranir:

```text
Resume file found: downloads/Candidato-Ututo-2017-UL.iso.download (44.69 MB). Download will continue from where it left off.
Starting ISO download (1/1)...
Resuming download from 44.69 MB: downloads/Candidato-Ututo-2017-UL.iso.download
```

### 6. Resume kapat ve sifirdan baslat

```bash
./glance --download --no-resume --iso https://www.ututo.org/downloads/Candidato-Ututo-2017-UL.iso
```

### 7. ISO'yu belirli bir klasore indir

```bash
./glance --download --iso https://www.ututo.org/downloads/Candidato-Ututo-2017-UL.iso --output /srv/iso-cache
```

### 8. ISO'yu tam istedigin dosya yoluna indir

```bash
./glance --download --iso https://www.ututo.org/downloads/Candidato-Ututo-2017-UL.iso --output-path /srv/releases/custom-ututo.iso
```

Not:

- `--output-path` sadece tek ISO icin gecerlidir
- Coklu secimde `--output` kullanilmalidir

### 9. Lokal ISO'yu kopyala ve checksum hesapla

```bash
./glance --download --iso /home/user/isos/archlinux-x86_64.iso
```

### 10. Belirli checksum ile dogrula

```bash
./glance --download --iso https://releases.ubuntu.com/24.04/ubuntu-24.04.4-live-server-amd64.iso --checksum e907d92eeec9df64163a7e454cbc8d7755e8ddc7ed42f99dbc80c40f1a138433
```

## Sistem Yonetici Rehberi

### Mirror ve depo kullanim onerileri

- FTP tarafinda anonim giris desteklenmeyen hostlarda HTTP/HTTPS taramasi tercih edilmelidir
- Buyuk mirror'larda tarama sirasinda rate limit uygulanir
- FTP taramada dizin istekleri arasinda gecikme kullanilir; bu sayede saldirgan trafik gibi gorunme riski azaltilir

### SSH/SFTP yukleme

Sifre ile:

```bash
./glance --upload --file ./downloads/ubuntu-24.04.4-live-server-amd64.iso --host 192.168.1.50 --user root --password secret
```

SSH key ile:

```bash
./glance --upload --file ./downloads/ubuntu-24.04.4-live-server-amd64.iso --host 192.168.1.50 --user root --ssh-key ~/.ssh/id_rsa --known-hosts ~/.ssh/known_hosts
```

Tek komutta indir ve yukle:

```bash
./glance --download --iso https://releases.ubuntu.com/24.04/ubuntu-24.04.4-live-server-amd64.iso --upload --host 192.168.1.50 --user root --ssh-key ~/.ssh/id_rsa --known-hosts ~/.ssh/known_hosts
```

`known_hosts` kaydi yoksa once ekleyin:

```bash
ssh-keyscan -H 192.168.1.50 >> ~/.ssh/known_hosts
```

### Operasyonel notlar

- Resume dosyalari final dosya adinin sonuna `.download` eklenerek tutulur
- Download tamamlaninca gecici dosya final dosyaya cevrilir
- Sunucu `Range` desteklemiyorsa arac otomatik olarak sifirdan yeniden indirir

## Gelistirici Rehberi

### Paket yapisi

- `internal/cli`: flag parse etme, prompt, secim ve akisin orkestrasyonu
- `internal/downloader`: HTTP/FTP tarama, indirme, resume ve checksum kaynak cozumu
- `internal/verifier`: lokal checksum hesaplama ve dogrulama
- `internal/uploader`: SSH/SFTP yukleme ve uzak checksum dogrulama
- `internal/keygen`: SSH anahtar uretimi

### Onerilen gelistirme komutlari

```bash
go build -o glance .
./glance --help
./glance --download --iso https://www.ututo.org/downloads/
```

### Checksum cozme sirasi

HTTP/HTTPS icin:

- `dosya.iso.sha256sum` / `dosya.iso.sha512sum` / `dosya.iso.md5sum`
- `SHA256SUMS` / `SHA512SUMS` / `MD5SUMS`
- `checksum`

FTP icin:

- ayni dizindeki benzer checksum dosyalari
- indeks dosyalari ve yaygin checksum dosya adlari

Bulunamazsa lokal hash hesaplanir.

### Interaktif karar akisi

1. Kaynak bir dizinse ISO listesi cikarilir
2. Kullanici bir veya daha fazla ISO secer
3. `--output` ya da `--output-path` verilmediyse hedef dizin sorulur
4. `.download` dosyasi varsa resume bilgisi gosterilir
5. Indirme biterse checksum dogrulama yapilir
6. `--upload` varsa uzak sisteme aktarilir

## Tum Flag'ler

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

## Changelog Notu

`ChangeLog.md` release notlari ve gunluk degisim kaydi icin tutulur. En guncel commit zamanini dogrulamak icin:

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
