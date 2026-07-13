#!/usr/bin/env python3
"""Regenerates internal/racing/catalog_seed.json from my-racing-planner's data
exports (real iRacing ids, names, prices, free flags, and the current season's
weekly schedules).

Usage:
    python3 gen-catalog-seed.py <dir-with-mrp-{series,cars,tracks}.json>

The season window is derived from the data itself: the most common first-week
date among 13-week series is week 1; every series' races are then slotted into
that 13-week grid by date (long, year-round series simply show the races that
fall inside the window). Descriptions from the previous seed are carried over
by name so hand-written copy survives regeneration.
"""

import json
import sys
import unicodedata
from collections import Counter
from datetime import date, timedelta
from pathlib import Path

HERE = Path(__file__).resolve().parent
SEED = HERE.parent / "internal" / "racing" / "catalog_seed.json"

CATEGORY_IDS = {"oval": 1, "road": 2, "dirt_oval": 3, "dirt_road": 4, "sports_car": 5, "formula_car": 6}
SEASON_WEEKS = 13


def norm(s: str) -> str:
    s = unicodedata.normalize("NFKD", s)
    return "".join(c for c in s.lower() if c.isalnum())


def main(src: Path) -> None:
    series = json.load(open(src / "mrp-series.json"))
    cars = json.load(open(src / "mrp-cars.json"))
    tracks = json.load(open(src / "mrp-tracks.json"))

    # Carry hand-written descriptions over from the current seed, keyed by name.
    old_car_desc, old_track_desc, old_series_desc = {}, {}, {}
    if SEED.exists():
        old = json.load(open(SEED))
        old_car_desc = {norm(c["carName"]): c.get("description", "") for c in old.get("cars", [])}
        old_track_desc = {norm(t["trackName"]): t.get("description", "") for t in old.get("tracks", [])}
        old_series_desc = {norm(s["seriesName"]): s.get("description", "") for s in old.get("series", [])}

    out_cars = [
        {
            "carId": c["id"],
            "carName": c["name"],
            "category": (c["categories"] or ["road"])[0],
            "description": old_car_desc.get(norm(c["name"]), ""),
            "free": bool(c["free"]),
            "price": c.get("price", 0),
        }
        for c in sorted(cars.values(), key=lambda c: c["id"])
    ]

    out_tracks = [
        {
            "trackId": t["id"],
            "trackName": t["name"],
            "configName": t.get("config", "") or "",
            "category": (t["categories"] or ["road"])[0],
            "description": old_track_desc.get(norm(t["name"]), ""),
            "free": bool(t["free"]),
            "price": t.get("price", 0),
            "skuGroup": t.get("sku", 0),
        }
        for t in sorted(tracks.values(), key=lambda t: t["id"])
    ]

    # Combined layouts are not purchasable directly (sku 0, not free): owning
    # every component purchase unlocks them. Model Nürburgring Combined as
    # requiring one Nordschleife config + one GP config (sku-group unlock covers
    # the specific configs the user marked).
    by_name = {}
    for t in tracks.values():
        by_name.setdefault(t["name"], []).append(t)
    requirements = {}
    for t in tracks.values():
        if t.get("sku", 0) == 0 and not t["free"] and t["name"] == "Nürburgring Combined":
            nord = min(x["id"] for x in by_name.get("Nürburgring Nordschleife", []) or [t])
            gp = min(x["id"] for x in by_name.get("Nürburgring Grand-Prix-Strecke", []) or [t])
            requirements[str(t["id"])] = [nord, gp]
    leftovers = [
        f'{t["id"]} {t["name"]} ({t.get("config")})'
        for t in tracks.values()
        if t.get("sku", 0) == 0 and not t["free"] and t["name"] != "Nürburgring Combined"
    ]
    if leftovers:
        print("NOTE: unpurchasable tracks with no requirements mapping:", leftovers, file=sys.stderr)

    # Season window: the modal week-1 date among 13-week series.
    starts = Counter(
        s["weeks"][0]["date"] for s in series.values() if len(s["weeks"]) == SEASON_WEEKS and s["weeks"]
    )
    season_start = date.fromisoformat(starts.most_common(1)[0][0])
    season_end = season_start + timedelta(weeks=SEASON_WEEKS)
    print(f"season window: {season_start} → {season_end} ({starts.most_common(3)})", file=sys.stderr)

    out_series = []
    scheduled = 0
    for s in sorted(series.values(), key=lambda s: s["id"]):
        weeks = []
        for w in s["weeks"]:
            d = date.fromisoformat(w["date"])
            if season_start <= d < season_end:
                weeks.append(
                    {
                        "week": (d - season_start).days // 7 + 1,
                        "trackId": w["track"]["id"],
                        "date": w["date"],
                    }
                )
        scheduled += len(weeks)
        out_series.append(
            {
                "seriesId": s["id"],
                "seriesName": s["name"],
                "category": s["category"],
                "categoryId": CATEGORY_IDS.get(s["category"], 0),
                "licenseNeeded": s["license"]["letter"],
                "description": old_series_desc.get(norm(s["name"]), ""),
                "cars": s.get("cars", []),
                "weeks": weeks,
            }
        )

    seed = {
        "seasonStart": season_start.isoformat(),
        "cars": out_cars,
        "tracks": out_tracks,
        "series": out_series,
        "trackRequirements": requirements,
    }
    SEED.write_text(json.dumps(seed, ensure_ascii=False, indent=1) + "\n")
    print(
        f"wrote {SEED}: {len(out_cars)} cars, {len(out_tracks)} tracks, "
        f"{len(out_series)} series, {scheduled} scheduled races, {len(requirements)} requirement rows",
        file=sys.stderr,
    )


if __name__ == "__main__":
    main(Path(sys.argv[1]))
