#!/usr/bin/env python3
"""
extract_spl.py
Parst alle *.csv im aktuellen Verzeichnis (rekursiv optional per --recursive)
und extrahiert aus SNMP-Zeilen Timestamp, IP sowie den SPL-Wert.

Korrektur: Der SPL steht als ASCII-Ziffern *zwischen* den ASCII-Klammern '[' (91)
und ']' (93) im OctetString. Beispiel:
... 32 91 54 48 93
=> 54->'6', 48->'0' => "60" => SPL=60
"""

import os
import re
import sys
import glob
import csv
from typing import Iterable, Tuple, Optional, List

def token_to_int(tok: str) -> Optional[int]:
    tok = tok.strip().strip(",]")
    if tok.lower().startswith("0x"):
        try:
            return int(tok, 16)
        except Exception:
            return None
    if re.fullmatch(r"[0-9]+", tok):
        try:
            return int(tok, 10)
        except Exception:
            return None
    if re.fullmatch(r"[0-9a-fA-F]+", tok):
        try:
            return int(tok, 16)
        except Exception:
            return None
    return None

def token_to_char(tok: str) -> Optional[str]:
    v = token_to_int(tok)
    if v is None:
        return None
    try:
        return chr(v)
    except Exception:
        return None

def extract_digits_from_tokens(tokens: List[str]) -> Optional[int]:
    """
    Bevorzugt die Zeichen zwischen ASCII '[' (91) und ']' (93).
    Falls kein solches Paar existiert, fällt auf die letzten beiden Tokens zurück.
    """
    ints = [token_to_int(t) for t in tokens]
    # Letztes '[' und anschließendes ']'
    open_idx = None
    close_idx = None
    for i in range(len(ints)-1, -1, -1):
        if ints[i] == 91:
            open_idx = i
            break
    if open_idx is not None:
        for j in range(open_idx + 1, len(ints)):
            if ints[j] == 93:
                close_idx = j
                break

    chars = None
    if open_idx is not None and close_idx is not None and close_idx > open_idx + 1:
        chars = [token_to_char(t) for t in tokens[open_idx+1:close_idx]]
    else:
        # Fallback: letzte zwei Tokens
        if len(tokens) >= 2:
            chars = [token_to_char(t) for t in tokens[-2:]]

    if not chars or any(c is None for c in chars):
        return None

    digits = "".join(ch for ch in "".join(chars) if ch.isdigit())
    if digits == "":
        return None
    try:
        return int(digits, 10)
    except Exception:
        return None

def parse_line(line: str) -> Optional[Tuple[str, str, int]]:
    parts = line.rstrip("\n").split(",", 4)
    if len(parts) < 5:
        return None

    ts_raw, ip_port, oid, typ, val = parts

    if typ.strip().lower() != "octetstring":
        return None

    # IP ohne Port (IPv6 in eckigen Klammern wird unterstützt)
    ip = ip_port.strip()
    if ip.startswith("[") and "]" in ip:
        ip_only = ip[1:ip.index("]")]
    else:
        ip_only = ip.split(":")[0]

    # Bytes innerhalb der eckigen Klammern des Feldes
    m = re.search(r"\[(.*?)\]", val)
    if not m:
        return None

    token_str = m.group(1).strip()
    if not token_str:
        return None

    tokens = [t for t in re.split(r"\s+", token_str) if t]
    if len(tokens) < 2:
        return None

    spl = extract_digits_from_tokens(tokens)
    if spl is None:
        return None

    timestamp = ts_raw.strip()
    return (timestamp, ip_only, spl)

def process_files(patterns: Iterable[str], out_path: str) -> int:
    rows = []
    for pattern in patterns:
        for path in glob.glob(pattern, recursive=True):
            if os.path.abspath(path) == os.path.abspath(out_path):
                continue
            try:
                with open(path, "r", encoding="utf-8", errors="replace") as f:
                    for line in f:
                        if line.lower().startswith("timestamp,") or line.lower().startswith("zeitstempel,"):
                            continue
                        parsed = parse_line(line)
                        if parsed is not None:
                            rows.append(parsed)
            except FileNotFoundError:
                continue

    rows.sort(key=lambda r: r[0])
    with open(out_path, "w", newline="", encoding="utf-8") as out:
        w = csv.writer(out)
        w.writerow(["timestamp", "ip", "spl"])
        for r in rows:
            w.writerow(r)
    return len(rows)

def main(argv: list[str]) -> int:
    out_path = "spl_extracted.csv"
    recursive = False
    patterns = ["*.csv"]

    i = 1
    while i < len(argv):
        a = argv[i]
        if a == "--out" and i + 1 < len(argv):
            out_path = argv[i + 1]
            i += 2
        elif a == "--recursive":
            recursive = True
            i += 1
        elif a == "--pattern" and i + 1 < len(argv):
            patterns = [argv[i + 1]]
            i += 2
        else:
            print(f"Unbekanntes Argument ignoriert: {a}")
            i += 1

    if recursive:
        if patterns == ["*.csv"]:
            patterns = ["**/*.csv"]

    count = process_files(patterns, out_path)
    print(f"Fertig. {count} Zeilen extrahiert -> {out_path}")
    return 0

if __name__ == "__main__":
    raise SystemExit(main(sys.argv))
