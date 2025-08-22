# SNMP Trap Receiver

Empfängt SNMP Traps und schreibt diese samt Inhalt in CSV-Dateien.

Läuft als Windows-Dienst `SNMPTrapReceiver`.

Logs werden in `service.log` geschrieben.

## Installation:

- Archiv entpacken und Ordner an einem fixen Ort ablegen, z.B. `C:\Program Files\`
- Terminal als Admin öffnen und in den Ordner der EXE navigieren
- Dienst installieren mit `.\snmptrap_service.exe install`
- Dienst starten mit `net start SNMPTrapReceiver`


## Konfiguration per config.ini (optional):

- Port
- Pfad, wo die CSVs abgelegt werden
- Rotation der CSV Files


## Deinstallation:

- Terminal als Admin öffnen und in den Ordner der EXE navigieren
- Dienst stoppen mit `net stop SNMPTrapReceiver`
- Dienst löschen mit `.\snmptrap_service.exe uninstall`

---

## Optional: SPL-Werte extrahieren mit `extract_spl.exe`

`extract_spl.exe` im Ordner mit den generierten CSV-Files starten. Es wird eine neue CSV mit den konsolidierten und bereinigten Daten generiert.