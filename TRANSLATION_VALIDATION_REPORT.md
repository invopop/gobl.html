# Translation Validation Report

Generated: 2025-10-27

## Executive Summary

This report validates all translation files in the project against the English base translations. The project supports **12 languages** with translations stored in YAML format.

**Key Findings:**
- **3 languages** (Catalan, Basque, Galician) have complete and up-to-date translations for `app.yml`
- **7 languages** have missing translation keys in `app.yml` (ranging from 34 to 88 missing keys)
- Several languages have minor discrepancies in `countries.yml`
- All non-English languages include `currencies.yml` files (English intentionally doesn't have this file)

---

## English Base Statistics

The English base translations consist of:
- **app.yml**: 238 keys (main application translations)
- **countries.yml**: 249 keys (ISO 3166-1 alpha-2 country codes)
- **currencies.yml**: Not present (intentionally omitted)

---

## Language-by-Language Analysis

### ✅ Complete Translations

The following languages have **complete** `app.yml` translations:

1. **Catalan (ca)** - 238/238 keys ✓
2. **Basque (eu)** - 238/238 keys ✓
3. **Galician (gl)** - 238/238 keys ✓

**Note:** These three languages were recently added (commit: "Translations for Catalan, Galician, and Basque") and are fully up-to-date.

**Minor observation:** All three have 14 extra country codes in `countries.yml` (non-standard ISO codes like AC, CP, DG, EA, EU, EZ, FX, IC, SU, TA, UK, UN, XK) - this is acceptable and provides better coverage.

---

### ⚠️ Languages with Missing Translations

#### 1. German (de)
- **app.yml**: Missing **84 keys** (65% complete)
- **countries.yml**: Missing 1 key

**Critical missing keys include:**
- Mexican CFDI extensions (`mx-cfdi-issue-place`, `mx-cfdi-payment-method`, `mx-cfdi-fiscal-regime`, etc.)
- Adjustment invoice title (`billing.invoice.title.adjustment`)
- Identity labels (`it-fiscal-code`, `de-tax-number`)
- Person-related fields (`org.party.person`, `org.party.person_label`)
- Portuguese regime titles (multiple)
- Various payment and contract-related fields

---

#### 2. Greek (el)
- **app.yml**: Missing **34 keys** (86% complete)
- **countries.yml**: Complete ✓
- **Extra keys**: 1 obsolete key

**Missing keys include:**
- Cheque payment method (`billing.invoice.payment.instructions.methods.cheque`)
- Adjustment invoice title
- Identity labels
- Person fields
- Some Portuguese regime titles

---

#### 3. Spanish (es)
- **app.yml**: Missing **82 keys** (66% complete)
- **countries.yml**: Has 14 extra keys (same as ca/eu/gl)

**Missing keys include:**
- Due dates plural form (`billing.invoice.payment.terms.due_dates.other`)
- Mexican CFDI extensions
- Identity labels
- Person fields
- Website fields
- Multiple Portuguese regime titles
- Payment term keys

---

#### 4. French (fr)
- **app.yml**: Missing **88 keys** (63% complete)
- **countries.yml**: Complete ✓
- **Extra keys**: 4 obsolete keys

**Missing keys include:**
- Payment terms keys (advanced, due_date, end_of_month)
- Mexican CFDI extensions
- Adjustment invoice title
- Identity labels
- Person fields
- Multiple regime-specific titles

---

#### 5. Italian (it)
- **app.yml**: Missing **85 keys** (64% complete)
- **countries.yml**: Missing 1 key
- **currencies.yml**: Has 152 keys (14 fewer than other languages)

**Missing keys include:**
- Cheque payment method
- Mexican CFDI extensions
- Adjustment invoice title
- Identity labels (including its own `it-fiscal-code`)
- Person fields
- Multiple Portuguese regime titles

---

#### 6. Polish (pl)
- **app.yml**: Missing **88 keys** (63% complete)
- **countries.yml**: Missing 1 key
- **Extra keys**: 2 obsolete keys
- **currencies.yml**: Has 152 keys (14 fewer than other languages)

**Missing keys include:**
- Cheque payment method
- Due dates plural form
- Mexican CFDI extensions
- Adjustment invoice title
- Prices include tax label (`billing.invoice.totals.prices_include`)
- Colombian municipality extension (`co-dian-municipality`)
- Identity labels
- Person fields

---

#### 7. Portuguese (pt)
- **app.yml**: Missing **64 keys** (73% complete)
- **countries.yml**: Complete ✓
- **Extra keys**: 1 obsolete key

**Missing keys include:**
- Cheque payment method
- Due dates plural form
- Mexican CFDI extensions
- Adjustment invoice title
- Colombian and Mexican extensions
- Identity labels
- Person fields
- Various Portuguese regime titles (ironically missing from Portuguese translations)

---

## Common Missing Translation Patterns

Across all incomplete languages, the following categories of keys are frequently missing:

### 1. Mexican CFDI Extensions (Tax System)
- `billing.invoice.summary.ext_map.mx-cfdi-issue-place`
- `billing.invoice.summary.ext_map.mx-cfdi-payment-method`
- `org.party.ext_map.mx-cfdi-fiscal-regime`
- `org.party.ext_map.mx-cfdi-post-code`
- `org.party.ext_map.mx-cfdi-use`

### 2. Identity and Person Fields
- `org.party.identity_labels.it-fiscal-code`
- `org.party.identity_labels.de-tax-number`
- `org.party.person`
- `org.party.person_label`

### 3. New Invoice Types
- `billing.invoice.title.adjustment`

### 4. Payment Methods
- `billing.invoice.payment.instructions.methods.cheque`

### 5. Payment Terms
- `billing.invoice.payment.terms.due_dates.other` (plural form)
- `billing.invoice.payment.terms.keys.advanced`
- `billing.invoice.payment.terms.keys.due_date`
- `billing.invoice.payment.terms.keys.end_of_month`

### 6. Portuguese Regime Titles
- Multiple `regimes.pt.title.*` keys (CC, CM, CS, LD, RA, RP, RE, etc.)

### 7. Greek Regime Titles
- Multiple `regimes.gr.title.*` keys

---

## Recommendations

### Priority 1: Critical Updates (Missing Core Functionality)
Update the following languages to include the adjustment invoice title and person fields:
- German (de)
- Greek (el)
- Spanish (es)
- French (fr)
- Italian (it)
- Polish (pl)
- Portuguese (pt)

### Priority 2: Country-Specific Extensions
Ensure country-specific regime labels are translated:
- **Portuguese regime titles** should be added to all languages (especially Portuguese itself)
- **Mexican CFDI keys** should be translated for all languages
- **Greek regime titles** should be reviewed

### Priority 3: Payment and Financial Terms
Update payment-related translations:
- Add cheque payment method translations
- Add plural forms for due dates
- Complete payment terms translations

### Priority 4: Minor Cleanup
- Remove obsolete extra keys from Greek (1), Spanish (1), French (4), Polish (2), Portuguese (1)
- Add missing country key to German, Italian, and Polish `countries.yml`

---

## Translation Completeness Matrix

| Language | app.yml | countries.yml | currencies.yml | Overall Status |
|----------|---------|---------------|----------------|----------------|
| ca (Catalan) | ✅ 100% (238/238) | ✅ 249 + 14 extra | ✅ 166 keys | Complete |
| de (German) | ⚠️ 65% (154/238) | ⚠️ 248/249 | ✅ 166 keys | Needs update |
| el (Greek) | ⚠️ 86% (204/238) | ✅ 249/249 | ✅ 166 keys | Mostly complete |
| es (Spanish) | ⚠️ 66% (156/238) | ✅ 249 + 14 extra | ✅ 166 keys | Needs update |
| eu (Basque) | ✅ 100% (238/238) | ✅ 249 + 14 extra | ✅ 166 keys | Complete |
| fr (French) | ⚠️ 63% (150/238) | ✅ 249/249 | ✅ 166 keys | Needs update |
| gl (Galician) | ✅ 100% (238/238) | ✅ 249 + 14 extra | ✅ 166 keys | Complete |
| it (Italian) | ⚠️ 64% (153/238) | ⚠️ 248/249 | ⚠️ 152 keys | Needs update |
| pl (Polish) | ⚠️ 63% (150/238) | ⚠️ 248/249 | ⚠️ 152 keys | Needs update |
| pt (Portuguese) | ⚠️ 73% (174/238) | ✅ 249/249 | ✅ 166 keys | Needs update |

---

## Next Steps

1. **Sync with translation service**: If using Localazy (as indicated by `localazy.json`), trigger a sync to push the new English keys to translators
2. **Review missing keys**: Verify that all missing keys are intentional or need translation
3. **Update older translations**: Prioritize updating German, French, Spanish, Italian, and Polish translations
4. **Test recent translations**: Verify that Catalan, Basque, and Galician translations render correctly in the application

---

## Validation Script

The validation script (`validate_translations.sh`) is available in the project root and can be re-run at any time to check translation status:

```bash
./validate_translations.sh
```

This script:
- Extracts all keys from English base files
- Compares each language's keys against the base
- Reports missing, extra, and complete translations
- Provides detailed lists of missing keys for debugging
