// api/download.js
// Black Six Matrix – Sichere Download-API nach PayPal-Zahlung

const crypto = require('crypto');
const SECRET = process.env.DOWNLOAD_SECRET || 'black-six-matrix-secret-2026'; // SICHERES SECRET SETZEN!
const USED_TOKENS = new Set(); // In Produktion persistieren (Redis, DB, etc.)

// --- KORRIGIERTE DOWNLOAD_MAP MIT VOLLSTÄNDIGEN URLS ---
const DOWNLOAD_MAP = {
  'I1 – ZÜNDFUNKE': 'https://github.com/86102262471878081179/Black-Six-Matrix/releases/download/v1.0/I1-ZUENDFUNKE.zip',
  'I2 – GLIMMER': 'https://github.com/86102262471878081179/Black-Six-Matrix/releases/download/v1.0/I2-GLIMMER.zip',
  'I3 – FUNKE': 'https://github.com/86102262471878081179/Black-Six-Matrix/releases/download/v1.0/I3-FUNKE.zip',
  'I4 – FLAMME': 'https://github.com/86102262471878081179/Black-Six-Matrix/releases/download/v1.0/I4-FLAMME.zip',
  'I5 – INFERNO': 'https://github.com/86102262471878081179/Black-Six-Matrix/releases/download/v1.0/I5-INFERNO.zip',
  'M1 – BOOTCAMP': 'https://github.com/86102262471878081179/Black-Six-Matrix/releases/download/v1.0/M1-BOOTCAMP.zip',
  'M2 – FORTGESCHRITTEN': 'https://github.com/86102262471878081179/Black-Six-Matrix/releases/download/v1.0/M2-FORTGESCHRITTEN.zip',
  'M3 – AUDIT': 'https://github.com/86102262471878081179/Black-Six-Matrix/releases/download/v1.0/M3-AUDIT.zip',
  'M4 – LEGAL-TECH': 'https://github.com/86102262471878081179/Black-Six-Matrix/releases/download/v1.0/M4-LEGAL-TECH.zip',
  'M5 – MEISTER': 'https://github.com/86102262471878081179/Black-Six-Matrix/releases/do
