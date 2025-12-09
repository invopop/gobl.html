#!/bin/bash
# Validation script to check all translations against the English base

set -euo pipefail

LOCALES_DIR="/home/sam/workspace/invopop/gobl.html/locales"
TEMP_DIR=$(mktemp -d)

# Function to extract all keys from a YAML file (removing the top-level language code)
extract_keys() {
    local file=$1
    if [ ! -f "$file" ]; then
        return
    fi
    # Use yq to get all leaf keys (paths to actual values, not intermediate objects)
    # Remove the first element (language code) from the path
    yq eval '.. | select(tag != "!!map" and tag != "!!seq") | path | .[1:] | join(".")' "$file" | sort | uniq
}

echo "================================================================================"
echo "TRANSLATION VALIDATION REPORT"
echo "================================================================================"

# Extract English base keys
echo ""
echo "Extracting English base keys..."
extract_keys "$LOCALES_DIR/en/app.yml" > "$TEMP_DIR/en_app_keys.txt"
extract_keys "$LOCALES_DIR/en/countries.yml" > "$TEMP_DIR/en_countries_keys.txt"

EN_APP_COUNT=$(wc -l < "$TEMP_DIR/en_app_keys.txt" || echo 0)
EN_COUNTRIES_COUNT=$(wc -l < "$TEMP_DIR/en_countries_keys.txt" || echo 0)

echo "English base has:"
echo "  - app.yml: $EN_APP_COUNT keys"
echo "  - countries.yml: $EN_COUNTRIES_COUNT keys"
echo ""
echo "Note: English does not have a currencies.yml file"

# Process each language
echo ""
echo "================================================================================"
echo "CHECKING EACH LANGUAGE"
echo "================================================================================"

ISSUES_FOUND=0
LANGS_OK=()

for lang_dir in "$LOCALES_DIR"/*; do
    if [ ! -d "$lang_dir" ]; then
        continue
    fi

    lang=$(basename "$lang_dir")

    # Skip English
    if [ "$lang" = "en" ]; then
        continue
    fi

    echo ""
    echo "[$lang]"

    LANG_HAS_ISSUES=0

    # Check app.yml
    if [ ! -f "$lang_dir/app.yml" ]; then
        echo "  ✗ Missing app.yml"
        LANG_HAS_ISSUES=1
    else
        extract_keys "$lang_dir/app.yml" > "$TEMP_DIR/${lang}_app_keys.txt"
        LANG_APP_COUNT=$(wc -l < "$TEMP_DIR/${lang}_app_keys.txt" || echo 0)

        # Find missing keys
        comm -23 "$TEMP_DIR/en_app_keys.txt" "$TEMP_DIR/${lang}_app_keys.txt" > "$TEMP_DIR/${lang}_app_missing.txt"
        MISSING_APP=$(wc -l < "$TEMP_DIR/${lang}_app_missing.txt" || echo 0)

        # Find extra keys
        comm -13 "$TEMP_DIR/en_app_keys.txt" "$TEMP_DIR/${lang}_app_keys.txt" > "$TEMP_DIR/${lang}_app_extra.txt"
        EXTRA_APP=$(wc -l < "$TEMP_DIR/${lang}_app_extra.txt" || echo 0)

        if [ "$MISSING_APP" -gt 0 ]; then
            echo "  ✗ app.yml: Missing $MISSING_APP keys"
            echo "    First 10 missing keys:"
            head -10 "$TEMP_DIR/${lang}_app_missing.txt" | sed 's/^/      - /'
            if [ "$MISSING_APP" -gt 10 ]; then
                echo "      ... and $((MISSING_APP - 10)) more"
            fi
            LANG_HAS_ISSUES=1
        fi

        if [ "$EXTRA_APP" -gt 0 ]; then
            echo "  ! app.yml: Has $EXTRA_APP extra keys not in English"
            LANG_HAS_ISSUES=1
        fi

        if [ "$MISSING_APP" -eq 0 ] && [ "$EXTRA_APP" -eq 0 ]; then
            echo "  ✓ app.yml: Complete ($LANG_APP_COUNT keys)"
        fi
    fi

    # Check countries.yml
    if [ ! -f "$lang_dir/countries.yml" ]; then
        echo "  ✗ Missing countries.yml"
        LANG_HAS_ISSUES=1
    else
        extract_keys "$lang_dir/countries.yml" > "$TEMP_DIR/${lang}_countries_keys.txt"
        LANG_COUNTRIES_COUNT=$(wc -l < "$TEMP_DIR/${lang}_countries_keys.txt" || echo 0)

        # Find missing keys
        comm -23 "$TEMP_DIR/en_countries_keys.txt" "$TEMP_DIR/${lang}_countries_keys.txt" > "$TEMP_DIR/${lang}_countries_missing.txt"
        MISSING_COUNTRIES=$(wc -l < "$TEMP_DIR/${lang}_countries_missing.txt" || echo 0)

        # Find extra keys
        comm -13 "$TEMP_DIR/en_countries_keys.txt" "$TEMP_DIR/${lang}_countries_keys.txt" > "$TEMP_DIR/${lang}_countries_extra.txt"
        EXTRA_COUNTRIES=$(wc -l < "$TEMP_DIR/${lang}_countries_extra.txt" || echo 0)

        if [ "$MISSING_COUNTRIES" -gt 0 ]; then
            echo "  ✗ countries.yml: Missing $MISSING_COUNTRIES keys"
            LANG_HAS_ISSUES=1
        fi

        if [ "$EXTRA_COUNTRIES" -gt 0 ]; then
            echo "  ! countries.yml: Has $EXTRA_COUNTRIES extra keys"
            LANG_HAS_ISSUES=1
        fi

        if [ "$MISSING_COUNTRIES" -eq 0 ] && [ "$EXTRA_COUNTRIES" -eq 0 ]; then
            echo "  ✓ countries.yml: Complete ($LANG_COUNTRIES_COUNT keys)"
        fi
    fi

    # Check currencies.yml
    if [ -f "$lang_dir/currencies.yml" ]; then
        LANG_CURRENCIES_COUNT=$(extract_keys "$lang_dir/currencies.yml" | wc -l || echo 0)
        echo "  ℹ currencies.yml: Present ($LANG_CURRENCIES_COUNT keys) - English doesn't have this"
    else
        echo "  ℹ currencies.yml: Not present (English also doesn't have it)"
    fi

    if [ "$LANG_HAS_ISSUES" -eq 0 ]; then
        LANGS_OK+=("$lang")
    else
        ISSUES_FOUND=$((ISSUES_FOUND + 1))
    fi
done

echo ""
echo "================================================================================"
echo "SUMMARY"
echo "================================================================================"

TOTAL_LANGS=$(find "$LOCALES_DIR" -mindepth 1 -maxdepth 1 -type d | grep -v "/en$" | wc -l)
LANGS_OK_COUNT=${#LANGS_OK[@]}

echo "Total languages checked: $TOTAL_LANGS"
echo "Languages with issues: $ISSUES_FOUND"
echo "Languages without issues: $LANGS_OK_COUNT"

if [ "$LANGS_OK_COUNT" -gt 0 ]; then
    echo ""
    echo "Languages with complete translations:"
    for lang in "${LANGS_OK[@]}"; do
        echo "  ✓ $lang"
    done
fi

# Cleanup
rm -rf "$TEMP_DIR"

echo ""
