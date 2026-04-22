# glance-cli

Modular Go CLI araci.

Amac:
- Herhangi bir isletim sistemi ISO'sunu URL veya lokal path ile almak
- ISO dosyasini SSH/SFTP ile hedef sunucuya yuklemek
- MIT lisans metnini komuttan gostermek

## Ozellikler

- `--download`: Verilen ISO kaynagindan indirir veya lokal path'ten kopyalar
- `--iso`: ISO source URL/path
- `--checksum`: Indirme/kopyalama sonrasi hash dogrulamasi (zorunlu)
- `--checksum-algo`: sha256 (default), sha512, md5
- Indirme esnasinda canli yuzde, anlik/ortalama hiz, ag hizi ve kalan sure (ETA)
- `--upload`: ISO veya dosyayi uzak sunucuya yukler
- Sifre veya SSH key ile kimlik dogrulama
- `known_hosts` dosyasi ile SSH host key dogrulamasi
- Eksik `--host` ve `--user` bilgilerini klavyeden isteme
- `--license`: MIT lisansi ve copyright metni

## Derleme

```bash
go mod tidy
go build -o glance .
```

## Kullanim

Sadece ISO indir:

```bash
./glance --download --iso https://releases.ubuntu.com/24.04/ubuntu-24.04.4-live-server-amd64.iso --checksum e907d92eeec9df64163a7e454cbc8d7755e8ddc7ed42f99dbc80c40f1a138433
```

Belirli checksum ile dogrulayarak indir:

```bash
./glance --download --iso https://releases.ubuntu.com/24.04/ubuntu-24.04.4-live-server-amd64.iso --checksum e907d92eeec9df64163a7e454cbc8d7755e8ddc7ed42f99dbc80c40f1a138433
```

Lokal ISO path'ten kopyala:

```bash
./glance --download --iso /home/user/isos/archlinux-x86_64.iso
```

Lokal dosyayi sha512 ile dogrula:

```bash
./glance --download --iso /home/user/isos/archlinux-x86_64.iso --checksum <sha512sum> --checksum-algo sha512
```

Indirilen ISO dosyasini upload et (sifre ile):

```bash
./glance --upload --file ./downloads/ubuntu-24.04.4-live-server-amd64.iso --host 192.168.1.50 --user root --password secret
```

Indirilen ISO dosyasini upload et (SSH key ile):

```bash
./glance --upload --file ./downloads/ubuntu-24.04.4-live-server-amd64.iso --host 192.168.1.50 --user root --ssh-key ~/.ssh/id_rsa --known-hosts ~/.ssh/known_hosts
```

Tek komutta indir ve yukle:

```bash
./glance --download --iso https://releases.ubuntu.com/24.04/ubuntu-24.04.4-live-server-amd64.iso --checksum e907d92eeec9df64163a7e454cbc8d7755e8ddc7ed42f99dbc80c40f1a138433 --upload --host 192.168.1.50 --user root --ssh-key ~/.ssh/id_rsa --known-hosts ~/.ssh/known_hosts
```

Not: `--download` islemi checksum dogrulamasi olmadan tamamlanmaz. `--checksum` vermezsen komut senden interaktif olarak ister.

`known_hosts` kaydi yoksa once su komutla ekleyebilirsin:

```bash
ssh-keyscan -H 192.168.1.50 >> ~/.ssh/known_hosts
```

Lisans metnini goster:

```bash
./glance --license
```

Yardim:

```bash
./glance --help
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
