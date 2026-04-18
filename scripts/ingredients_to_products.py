#!/usr/bin/env python3
"""Extract unique food names from recipe ingredients and add as products."""
import re
import sqlite3

# Common units/quantities to strip from ingredient text.
UNITS = {
    'cup', 'cups', 'tablespoon', 'tablespoons', 'tbsp', 'teaspoon', 'teaspoons', 'tsp',
    'ounce', 'ounces', 'oz', 'pound', 'pounds', 'lb', 'lbs', 'gram', 'grams', 'g', 'kg',
    'quart', 'quarts', 'qt', 'gallon', 'gallons', 'gal', 'liter', 'liters',
    'pinch', 'dash', 'bunch', 'clove', 'cloves', 'slice', 'slices', 'piece', 'pieces',
    'can', 'cans', 'package', 'packages', 'bag', 'bags', 'bottle', 'bottles',
    'jar', 'jars', 'box', 'boxes', 'carton', 'cartons', 'container', 'stick', 'sticks',
    'head', 'heads', 'sprig', 'sprigs', 'handful', 'large', 'medium', 'small',
    'inch', 'inches',
}

def extract_food_name(text):
    """Try to extract the food name from an ingredient string."""
    if not text:
        return None

    # Remove parenthetical notes.
    text = re.sub(r'\([^)]*\)', '', text)

    # Remove leading quantities: numbers, fractions, ranges.
    text = re.sub(r'^[\d\s/.\-–]+', '', text.strip())

    # Remove leading unit words.
    words = text.split()
    while words and words[0].lower().rstrip('.,;:') in UNITS:
        words.pop(0)

    # Remove "of" at the start.
    if words and words[0].lower() == 'of':
        words.pop(0)

    name = ' '.join(words).strip()

    # Remove trailing prep instructions after comma.
    name = re.split(r',\s', name)[0]

    # Clean up.
    name = name.strip(' ,;:.-')
    name = re.sub(r'\s+', ' ', name)

    # Skip if too short or looks like a note.
    if len(name) < 2:
        return None
    if name.lower().startswith(('to taste', 'as needed', 'optional', 'for ', 'freshly')):
        return None

    # Title case.
    return name.title()

def main():
    db = sqlite3.connect('data/domicile.db')
    cur = db.cursor()

    # Get all ingredient texts.
    cur.execute("SELECT DISTINCT text FROM recipe_ingredients")
    ingredients = [row[0] for row in cur.fetchall()]

    # Get existing product names (lowercase for comparison).
    cur.execute("SELECT name FROM products")
    existing = {row[0].lower() for row in cur.fetchall()}

    # Get default location and QU.
    cur.execute("SELECT id FROM locations LIMIT 1")
    loc_row = cur.fetchone()
    location_id = loc_row[0] if loc_row else 1

    cur.execute("SELECT id FROM quantity_units LIMIT 1")
    qu_row = cur.fetchone()
    qu_id = qu_row[0] if qu_row else 1

    # Extract unique food names.
    foods = set()
    for text in ingredients:
        name = extract_food_name(text)
        if name and name.lower() not in existing and name.lower() not in foods:
            foods.add(name.lower())

    # Insert as products.
    added = 0
    for food in sorted(foods):
        name = food.title()
        # Skip very generic things.
        if name.lower() in ('water', 'ice', 'salt', 'pepper', 'oil', 'sugar', 'flour'):
            continue
        if len(name) < 3:
            continue

        cur.execute("""
            INSERT OR IGNORE INTO products (name, location_id, qu_id_purchase, qu_id_stock, qu_factor_purchase_to_stock)
            VALUES (?, ?, ?, ?, 1.0)
        """, (name, location_id, qu_id, qu_id))
        if cur.rowcount > 0:
            added += 1

    db.commit()
    db.close()
    print(f"Extracted {len(foods)} unique food names, added {added} new products")

if __name__ == '__main__':
    main()
