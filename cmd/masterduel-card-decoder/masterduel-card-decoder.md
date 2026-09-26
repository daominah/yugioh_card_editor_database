# Master Duel Card Decoder

A process for guessing the hidden monster in the Master Duel card decoder game,
using `data/yugioh.db` to pick the guess that leaves the fewest candidates.

## Game rules

- The game **hides one monster** and **reveals 1 of its stats** at the start.
- The player picks a card as a **guess**.
  The game compares the guessed card with the hidden monster
  and reports each stat as **match or unmatch**.
  It also reports how many **candidates remain**.
- Instead of a guess, the player can request a **hint**,
  which **confirms 1 more stat** of the hidden monster, picked by the game.
- The player keeps guessing, narrowing down with the revealed stats,
  until a guess **matches all 6 stats**.

## How to play

Stats: Attribute, Frame, Type, Level, ATK, DEF.
Goal: each guess **leaves as few candidates** as possible, whatever the game reports.

- Candidates: monsters in `data/yugioh.db` with **every confirmed value**
  and **none of the excluded values** (unmatched guesses pile up).
- Rank: rate each candidate by **how well it splits the rest**,
  as in "Rank next guesses" below.
- Next guess: list up to **8 top-ranked candidates**,
  in a second table too when the 2 ranking orders disagree.
  Cards with identical stats collapse into 1 row, led by the smallest card ID.
  Show card names; add the card ID only when needed to clarify.
- Next action: the script's last line, "REQUEST A HINT" or "GUESS" a card.
  Put it last in the round notes, under "Next guess", and last in the reply,
  where the player looks.
- Hint: record it as a "Hint" row in the guesses table and add the value to Confirmed.
- Check: the player **compares the game's candidates count with the database's**.
  The game's should be a bit lower, because the game lacks some new cards.
  A big gap means a wrong filter.
- Notes: after each guess or hint, **record the round**
  in `cmd/masterduel-card-decoder/tmp-masterduel-round.md` as in the example below.
  Delete the file when the player says the round is won.

## Query candidates

Edit `confirmed` and `excluded` at the top of `masterduel_card_decoder.go`,
then run `go run ./cmd/masterduel-card-decoder` to query `data/yugioh.db`.

The game treats Pendulum as its own frame,
but the database stores Pendulum monsters as `MonsterEffect` (or another frame)
with `is_pendulum = 1`.
So Effect means `card_subtype = 'MonsterEffect' AND is_pendulum = 0`.

## Rank next guesses

Try each candidate as the guess against each candidate as the hidden monster,
and group the hidden monsters by the match pattern the game would report.
Candidates in the same group cannot be told apart by that guess.

- Average left: the number of cards left, averaged over every candidate as the hidden monster.
  A win leaves 0.
  Ranking by this first leaves the fewest cards on average, accepting a worse unlucky result.
- Worst case: the largest group left after the guess.

Neither order is proven better, and they usually agree.
Rank by average left first, with worst case breaking ties.
When the other order puts a different guess on top,
**show both tables** and let the player pick.

Only candidates are tried as guesses.

The duration grows with the square of the candidates count:
about 2 seconds for 2600 candidates, 24 seconds for the whole pool of 9275.
Over 3000 candidates the script warns before the check,
which only happens when the start reveals Frame Effect (5913 candidates).

Example: 4 candidates with Dinosaur and ATK 2000 confirmed,
EARTH, Effect, Level 4, and DEF 0 excluded.

| Card                    | Attribute | Frame   | Level |  DEF | Average left | Worst case |
|:------------------------|:----------|:--------|------:|-----:|-------------:|-----------:|
| Grenosaurus             | FIRE      | Xyz     |     3 | 1900 |         0.75 |          1 |
| Horned Brute Carnovorus | FIRE      | Synchro |     6 |  200 |         0.75 |          1 |
| Number 19: Freezadon    | WATER     | Xyz     |     5 | 2500 |         1.25 |          2 |
| Horned Saurus           | DARK      | Fusion  |     6 | 1800 |         1.25 |          2 |

Guessing "Number 19: Freezadon", "Horned Brute Carnovorus" and "Horned Saurus"
both report nothing matched, so 2 cards can be left:
the average is (0 win + 1 "Grenosaurus" + 2 + 2) / 4 = 1.25.
Guessing "Grenosaurus", every hidden monster reports a different pattern:
(0 + 1 + 1 + 1) / 4 = 0.75.

## Compare hints

A hint leaves the candidates sharing the hidden monster's value of 1 stat.
The game picks the stat, so the "random" row averages over the unconfirmed stats.

A hint is scarce, so recommend it only when **clearly better** than the best guess.
A hint leaving 100 cards on average against a guess leaving 120 is not a clear winner.
A clear winner example: with Level 8, DARK, and Effect confirmed,
among 93 candidates, a hint leaves 9.62 on average and the best guess leaves 28.39,
so the next action is "REQUEST A HINT".

## How to note after each guess

Example: a Dinosaur round right after submitting the first guess, "Flowerdino".
The second guess, "Grenosaurus", matched all 6 stats.

### Current state

- Confirmed: Dinosaur, ATK 2000.
- Excluded:
  - Attribute: EARTH.
  - Frame: Effect.
  - Level: 4.
  - DEF: 0.
- Remaining candidates: 4.

### Remaining candidates stats

| Stat      | Top values (candidates)               |
|:----------|:--------------------------------------|
| Attribute | FIRE (2), DARK (1), WATER (1)         |
| Frame     | Xyz (2), Fusion (1), Synchro (1)      |
| Level     | 6 (2), 3 (1), 5 (1)                   |
| DEF       | 1800 (1), 1900 (1), 200 (1), 2500 (1) |

### Guesses

| Guess | Card       | Matched            | Candidates after | Unmatched                     |
|:------|:-----------|:-------------------|-----------------:|:------------------------------|
| 1     | Flowerdino | Dinosaur, ATK 2000 |                4 | EARTH, Effect, Level 4, DEF 0 |

### Hints

| Hinted stat | Average left | Worst case |
|:------------|-------------:|-----------:|
| Attribute   |         1.50 |          2 |
| Frame       |         1.50 |          2 |
| Level       |         1.50 |          2 |
| DEF         |         1.00 |          1 |
| Random      |         1.38 |            |

### Next guess

| Suggestion              | Attribute | Frame   | Level |  DEF | Average left | Worst case |
|:------------------------|:----------|:--------|------:|-----:|-------------:|-----------:|
| Grenosaurus             | FIRE      | Xyz     |     3 | 1900 |         0.75 |          1 |
| Horned Brute Carnovorus | FIRE      | Synchro |     6 |  200 |         0.75 |          1 |
| Number 19: Freezadon    | WATER     | Xyz     |     5 | 2500 |         1.25 |          2 |
| Horned Saurus           | DARK      | Fusion  |     6 | 1800 |         1.25 |          2 |

**GUESS Grenosaurus**: leaves 0.75 on average, a hint leaves 1.38.
