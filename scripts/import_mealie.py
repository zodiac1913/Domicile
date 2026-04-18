#!/usr/bin/env python3
"""Import recipes from Mealie database.json into Domicile SQLite."""
import json
import sqlite3
import re

def slugify(text):
    text = text.lower().strip()
    text = re.sub(r'[^\w\s-]', '', text)
    text = re.sub(r'[\s_]+', '-', text)
    text = re.sub(r'-+', '-', text)
    return text.strip('-')

def main():
    with open('data/database.json') as f:
        data = json.load(f)

    db = sqlite3.connect('data/domicile.db')
    cur = db.cursor()

    # Build lookup maps
    categories = {c['id']: c['name'] for c in data.get('categories', [])}
    tags = {t['id']: t['name'] for t in data.get('tags', [])}
    r2c = {}  # recipe_id -> [category names]
    for link in data.get('recipes_to_categories', []):
        rid = link.get('recipe_id')
        cid = link.get('category_id')
        if rid and cid and cid in categories:
            r2c.setdefault(rid, []).append(categories[cid])
    r2t = {}  # recipe_id -> [tag names]
    for link in data.get('recipes_to_tags', []):
        rid = link.get('recipe_id')
        tid = link.get('tag_id')
        if rid and tid and tid in tags:
            r2t.setdefault(rid, []).append(tags[tid])

    imported = 0
    for r in data.get('recipes', []):
        name = r.get('name', '').strip()
        if not name:
            continue

        slug = r.get('slug') or slugify(name)
        desc = r.get('description', '')
        image = r.get('image', '')
        total_time = r.get('total_time', '')
        prep_time = r.get('prep_time', '')
        cook_time = r.get('cook_time', '')
        recipe_yield = r.get('recipe_yield', '')
        servings = r.get('recipe_servings', 1) or 1
        source_url = r.get('org_url', '')
        rating = r.get('rating')
        date_added = r.get('date_added', '')
        mealie_id = r['id']

        cur.execute("""
            INSERT INTO recipes (name, slug, description, image, total_time, prep_time,
                cook_time, recipe_yield, servings, source_url, rating, date_added)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        """, (name, slug, desc, image, total_time, prep_time,
              cook_time, recipe_yield, servings, source_url, rating, date_added))
        recipe_id = cur.lastrowid

        # Ingredients
        ingrs = [i for i in data.get('recipes_ingredients', []) if i.get('recipe_id') == mealie_id]
        ingrs.sort(key=lambda x: x.get('position', 0))
        for i in ingrs:
            text = i.get('original_text') or i.get('note') or i.get('note_normalized') or ''
            if not text.strip():
                continue
            title = i.get('title', '')
            pos = i.get('position', 0)
            cur.execute("INSERT INTO recipe_ingredients (recipe_id, position, text, title) VALUES (?, ?, ?, ?)",
                        (recipe_id, pos, text.strip(), title))

        # Instructions
        instrs = [i for i in data.get('recipe_instructions', []) if i.get('recipe_id') == mealie_id]
        instrs.sort(key=lambda x: x.get('position', 0))
        for i in instrs:
            text = i.get('text', '').strip()
            if not text:
                continue
            title = i.get('title', '')
            pos = i.get('position', 0)
            cur.execute("INSERT INTO recipe_instructions (recipe_id, position, title, text) VALUES (?, ?, ?, ?)",
                        (recipe_id, pos, title, text))

        # Categories
        for cat in r2c.get(mealie_id, []):
            cur.execute("INSERT OR IGNORE INTO recipe_categories (recipe_id, category) VALUES (?, ?)",
                        (recipe_id, cat))

        # Tags
        for tag in r2t.get(mealie_id, []):
            cur.execute("INSERT OR IGNORE INTO recipe_tags (recipe_id, tag) VALUES (?, ?)",
                        (recipe_id, tag))

        imported += 1

    db.commit()
    db.close()
    print(f"Imported {imported} recipes")

if __name__ == '__main__':
    main()
