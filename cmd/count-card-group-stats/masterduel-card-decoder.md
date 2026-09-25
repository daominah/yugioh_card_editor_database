# Master Duel Card Decoder

A process for guessing the hidden monster in the Master Duel card decoder game,
using `data/yugioh.db` to pick the guess whose stats are shared by the most candidates.

## Game rules

- The game **hides one monster** and **reveals 1 of its stats** at the start.
- The player picks a card as a **guess**.
  The game compares the guessed card with the hidden monster
  and reports each stat as **match or unmatch**.
  It also reports how many **candidates remain**.
- Instead of a guess, the player can request a **hint**,
  which **confirms 1 more stat** of the hidden monster.
- The player keeps guessing, narrowing down with the revealed stats,
  until a guess **matches all 6 stats**.

## How to play

Stats: Attribute, Frame, Type, Level, ATK, DEF.
Goal: each guess has **stats shared by as many candidates** as possible.

- Candidates: monsters in `data/yugioh.db` with **every confirmed value**
  and **none of the excluded values** (unmatched guesses pile up).
- Score: rate each candidate by **how popular its stats are** among the candidates,
  as in "Score remaining candidates" below.
- Next guess: list up to **8 top-scoring candidates**.
  Cards with identical stats collapse into 1 row, led by an **old, well-known card**.
  Show card names; add the card ID only when needed to clarify.
- Hint: once some stats are confirmed but **ATK or DEF is still not**,
  **remind the player to request a hint**.
  ATK and DEF have many values, so they are the hardest stats to guess.
  Record a hint as a "Hint" row in the guesses table and add the value to Confirmed.
- Check: the player **compares the game's candidates count with the database's**.
  The game's should be a bit lower, because the game lacks some new cards.
  A **big gap means a wrong filter**.
- Notes: after each guess, **record the round**
  in `cmd/count-card-group-stats/tmp-masterduel-round.md` as in the example below.
  Delete the file when the player says the round is won.

## Query candidates

Use a script similar to `cmd/read-yugiohdb-stats` to query `data/yugioh.db`.

The game treats Pendulum as its own frame,
but the database stores Pendulum monsters as `MonsterEffect` (or another frame)
with `is_pendulum = 1`.
So Effect means `card_subtype = 'MonsterEffect' AND is_pendulum = 0`.

## Score remaining candidates

For each candidate card and each unconfirmed stat,
count the candidates that share the card's value (the card itself included).
The score is the sum of these counts.

Scoring every card beats picking the most popular value stat by stat:
locking in the top Type first leaves ATK and DEF chosen among that Type only,
where values popular in the Type can be rare in the whole pool.

Example: 122 candidates with DARK, Effect, and Level 4 confirmed.

| Card                           | Type         |  ATK |  DEF | Score             | Same stats            |
|:-------------------------------|:-------------|-----:|-----:|:------------------|:----------------------|
| Machina Unclaspare             | Machine      | 1800 |  800 | 24 + 21 + 8 = 53  |                       |
| Genex Ally Duradark            | Machine      | 1800 |  200 | 24 + 21 + 4 = 49  |                       |
| Machina Possesstorage          | Machine      | 1600 | 1500 | 24 + 11 + 14 = 49 |                       |
| Gimmick Puppet Fiendish Knight | Machine      | 1800 |  500 | 24 + 21 + 4 = 49  |                       |
| Regenerating Mummy             | Zombie       | 1800 | 1500 | 12 + 21 + 14 = 47 |                       |
| Trap Reactor Y FI              | Machine      |  800 | 1800 | 24 + 12 + 11 = 47 |                       |
| Genex Ally Crusher             | Machine      | 1000 | 2000 | 24 + 15 + 8 = 47  | Time Thief Bezel Ship |
| Raidraptor - Tribute Lanius    | Winged Beast | 1800 |  400 | 18 + 21 + 6 = 45  |                       |

## How to note after each guess

Example: an Insect round right after submitting the second guess, Dragonbite.
The third guess, Splitting Planarian, matched all 6 stats.

### Current state

- Confirmed: DARK, Insect, Effect, Level 4, ATK 1000.
- Excluded:
  - DEF: 0, 1000.
- Remaining candidates: 1.

### Remaining candidates stats

| Stat | Top values (candidates) |
|:-----|:------------------------|
| DEF  | 800 (1)                 |

### Guesses

| Guess | Card        | Matched                                 | Unmatched             | Candidates after |
|:------|:------------|:----------------------------------------|:----------------------|-----------------:|
| 1     | Baby Spider | DARK, Insect, Effect                    | Level 3, ATK 0, DEF 0 |               15 |
| 2     | Dragonbite  | DARK, Insect, Effect, Level 4, ATK 1000 | DEF 1000              |                1 |

### Next guess

| Suggestion          | DEF | Score |
|:--------------------|----:|------:|
| Splitting Planarian | 800 |     1 |
