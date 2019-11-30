# craftberry-00 Setup & Konfiguration

Hier wird festgehalten, wie der Raspberry Pi craftberry-00 eingerichtet und konfiguriert wurde.

## Wo finde ich was?

| lokal     | extern     |
|-|-|
| 192.168.178.2 | https://home.siegelfabian.de | 
| craftberry-00 (Mit Avahi) ||
| pi.hole (Dank DNS) ||

| Dienst | lokal | Extern | 
|-|-|-|
| Home-Assistant | [IP:8123](http://192.168.178.2:8123) | [DOMAIN](https://home.siegelfabian.de) | 
| nextcloud | ❌ | [DOMAIN:8080](https://home.siegelfabian.de:8080) |
| pihole | [IP:8081](http://192.168.178.2:8123) | ❌|


## Hardware

 * Raspberry Pi 4 2GB + 32 GB SD-Karte
 * [Seagate Ironwolf 4TB ST4000VN008-2DR166](https://www.amazon.de/Seagate-ST4000VN008-IronWolf-interne-Festplatte/dp/B01LOOJBQY) in [LOGILINK UA0276](https://www.pollin.de/p/festplattengehaeuse-logilink-ua0276-3-5-usb-3-0-schraubenloses-design-703532)
 * 433Mhz Sender [FS1000A](https://cdn-reichelt.de/documents/datenblatt/FS1000A_DB_DE.pdf)

## Software
### Raspbian Konfiguration
#### `raspi-config`

```
2 Network-Options > hostname > `craftberry-00`
3 Boot Options > B1 Desktop CLI > B2 Commandline autologin
4 Localisation Options > Zeitzone & de-Locals generieren, Sprache trotzdem englisch
5 Interfacing Options > ssh
7 Advanced Options > Expand filesystem
```

#### Netzwerk

In der fritz!box dhcp-statische IP auf `192.168.178.2`

#### Externe Festplatte

in `/etc/fstab`:

```
UUID=14c26ec0-ea3a-4fc3-bebd-0d480b2390f0 /mnt/external ext4 defaults,auto,users,rw,nofail,x-systemd.device-timeout=30 0 0
```

#### Daily Reboot

Der Pi wird täglich um 2 Uhr neu gestartet:

in `/etc/crontab`:
```
# Example of job definition:
# .---------------- minute (0 - 59)
# |  .------------- hour (0 - 23)
# |  |  .---------- day of month (1 - 31)
# |  |  |  .------- month (1 - 12) OR jan,feb,mar,apr ...
# |  |  |  |  .---- day of week (0 - 6) (Sunday=0 or 7) OR sun,mon,tue,wed,thu,fri,sat
# |  |  |  |  |
# *  *  *  *  * user-name command to be executed

...

0  2    * * *   root    reboot
```

### zsh

[oh-my-zsh](https://github.com/robbyrussell/oh-my-zsh) mit `agnoster`-theme

### Samba

```bash
sudo apt install samba
```

Wir nutzen (diesmal) Unix-User, deren Home-Verzeichnisse auf der externen Festplatte liegen. Die Passwörter müssen allerdings mit `smbpasswd` festgelegt werden.

```bash
sudo useradd -m -d /mnt/external/shares/private/fabian fabian
sudo smbpasswd -a fabian
```

in `/etc/samba/smb.conf`:
```
[homes]
   comment = Home Directories
   browseable = no
   read only = no
   create mask = 0700
   directory mask = 0700
   valid users = %S

[public]
    comment=Public
    path=/mnt/external/shares/public
    writable=yes
    guest ok=yes
    public=yes
```

Beim Umzug von PC auf Raspberry-Pi gab es einen Missmatch zwischen GUIDs auf den Systemen. 
```
# Bei User-Shares
chown -R user:sambashare user/
# Bei public
chown -R nobody:sambashare public
```

### home-assistant

Manuelle Installation als [eigener Nutzer](https://www.home-assistant.io/docs/installation/virtualenv/) mit [systemd-unit](https://www.home-assistant.io/docs/autostart/systemd/#python-virtual-environment) (Achtung, zwecks Dummheit heißt die Unit `home-asstant@pi.service`)

**softlink**: `~/dot-homeassistant`

#### mosquitto

Der interne mqtt-broker von  home-assistant gilt als deprecated, deshalb nutzen wir nen eigenen mosquitto-broker 
```bash
sudo apt install mosquitto
```

#### [raspberry-remote](https://github.com/xkonni/raspberry-remote)

Wiring-Pi muss nicht manuell installiert werden. 

Um raspberry-remote zu kompilen ist es notwendig, [diesen Path](https://github.com/xkonni/raspberry-remote/pull/32) anzuwenden. (Stand: 06.10.19)

Da wir raspberry-remote im usermode bedienen, ist es notwenig, den gpio 17 zu exportieren. Wir machen das per `crontab -e`:
```crontab
@reboot gpio export 17 out

# mosquitto ist anscheinend ohne spezielle Konfig einfach public - dagegen kann man allerdings was machen
sudo mosquitto\_passwd -c /etc/mosquitto/passwd homeassistant

```
dann in `/etc/mosquitto/conf.d/00_fabian.conf`:
```conf
password_file /etc/mosquitto/passwd
allow_anonymous false
```

#### Luftdaten
In Hass-Web-Interface die Integration einrichtet

#### User
`fabian` und `juergen`

### LEMP
#### nginx

Für die verschiedenen Services sind je eigene configs/virtuelle server in `/etc/nginx/sites-available` angelegt, siehe config

#### PHP

```bash
sudo apt install php php-common php-fpm php-curl php-gd php-fpm php-cli php7.3-opcache php-json php-mbstring php-xml php-zip php-mysql php-imagick php-intl php-smbclient smbclient
```

Weitere Konfiguration bei [#nextcloud][nextcloud] 

#### MariaDB

```bash
sudo apt install mariadb-server
# aber als ob ne mariadb-installation funktioniert
sudo mysql -u root

[mysql] use mysql;
[mysql] update user set plugin='' where User='root';
[mysql] flush privileges;
[mysql] \q


mysql_secure_installation
```

[siehe](https://stackoverflow.com/a/47108020)

Es gibt **keine** phpmyadmin-Installation

#### SSL / certbot
```bash
sudo certbot --nginx
```

Certbot legt automatisch einen systemd timer zum revalidieren an.

### NextCloud
***softlink***: `~/www-nextcloud`

Installation via [zip-file](https://nextcloud.com/install/#instructions-server) in `/var/www/nextcloud`

Konfiguration siehe teils nginx-Config ([Vorlage](https://docs.nextcloud.com/server/15/admin_manual/installation/nginx.html)), Optimierungen (welche auch auf der Nextcloud selbst vorgeschlagen werden) nach:

 - [Tune php-fmp](https://docs.nextcloud.com/server/14/admin_manual/configuration_server/server_tuning.html#tune-php-fpm)
 - [Enable PHP OP-Cache](https://docs.nextcloud.com/server/14/admin_manual/configuration_server/server_tuning.html#tune-php-fpm)
 - "PHP-Memory-Cache konfiguriert" kann ignoriert werden

#### Samba-Integration:

Auf NextCloud die Erweiterung: "External Storages" herunterladen und einrichten.

#### Previews

Der PI ist zu schwach, um die Galerie-Bilder sinnvoll zu generieren. Deshalb lassen wir sie nur in 512px generieren, weil joa, ist halt (bisserl) weniger Arbeit.

In `www-nextcloud/config/config.php` hinzufügen:
```
  ...
  'preview_max_x' => 512,
  'preview_max_y' => 512,
  'preview_max_scale_factor' => 1,
  ...
```

#### E-Mail
Eingerichtet mit `home@siegelfabian.de`

#### User
`fabian` und `tim`

### pi-hole

Mit [Repository-Installer](https://github.com/pi-hole/pi-hole/#method-1-clone-our-repository-and-run) installiert. Dabei nicht lighthttp installieren.  Danach in `/etc/dhcpcd.conf` die Zeile auskommentiert:
```conf
        static ip_address=192.168.178.2/24
```

Passwort geändert mit `pihole -a -p`

FritzBox-Konfiguration geändert:
```
192.168.178.1 -> Heimnetz -> Netzwerk -> Netzwerkeinstellungen -> IP-Adressen # IPv4-Adressen -> Heimnetz # Lokaler DNS-Server
```

### motd

Veränderte Version von [gagle/raspberrypi-motd](https://github.com/gagle/raspberrypi-motd).
