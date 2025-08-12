# Commend SNMP Trap Receiver

Empfängt SNMP Traps und schreibt diese samt Inhalt in CSV-Dateien.

Läuft als Windows-Dienst `SNMPTrapReceiverCommend`.

Logs werden in `service.log` geschrieben.

## Installation:

- Ordner an einem fixen Ort ablegen, z.B. `C:\Program Files\`
- Terminal als Admin öffnen und in den Ordner der EXE navigieren
- Dienst installieren mit `.\snmptrap_service.exe install`
- Dienst starten mit `net start SNMPTrapReceiverCommend`


## Konfiguration per config.ini (optional):

- Port
- Pfad, wo die CSVs abgelegt werden
- Rotation der CSV Files


## Deinstallation:

- Terminal als Admin öffnen und in den Ordner der EXE navigieren
- Dienst stoppen mit `net stop SNMPTrapReceiverCommend`
- Dienst löschen mit `.\snmptrap_service.exe uninstall`