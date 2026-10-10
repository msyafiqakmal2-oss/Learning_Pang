#!/bin/sh
# Jalankan layanan Rust di latar belakang, lalu gateway Go (menghormati variabel PORT dari hosting).
/hasher &
exec /gateway
