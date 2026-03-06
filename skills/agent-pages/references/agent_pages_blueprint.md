# Agent Pages Blueprint

Use this reference when the claw needs help turning raw work into a public page with identity, continuity, and taste.

## Core framing

An agent page should communicate three things fast:

1. A recognizable character
2. A visible stream of work
3. A clear path to follow, copy, or adopt

The strongest pages feel like a living dossier rather than a product brochure.

## Static-site rule

If the site is deployed as a CDN static site, treat content as first-class pages inside the site:

- diary entries belong under routes like `/diary`
- works should have readable detail pages, not just teaser cards
- adopt/clone guidance should live inside the same site tree

Do not default to a platform marketing homepage when the goal is to publish one claw's public presence.

## Homepage block order

Use this order by default:

1. Name + one-line promise
2. Owner/operator identity
3. What the agent did in the last 7 days
4. Three best proof points
5. What this agent is good at
6. CTA: follow, adopt, clone, or contact

## Recommended homepage sections

### Hero

- Agent name
- Short role statement
- One sentence that captures taste or worldview
- Primary CTA

Example shape:

`三万是傅盛养出来的 AI 龙虾，负责连续工作、持续学习、把结果公开给世界看。`

### Recent wins

Show 3-5 concrete outcomes from the last week or month.

Good examples:

- Published 4 research notes
- Closed 12 monitoring alerts
- Shipped 2 new skills
- Ran 7x24 without manual takeover for 5 days

### Works

Feature outputs that prove capability:

- Reports
- Articles
- Dashboards
- Generated pages
- Workflows
- Screenshots or links

### Skills / stack

List what is reusable:

- Personal workflow packs
- Prompt or skill packs
- Tool chain
- Domain specializations

### Adoption CTA

Readers should know what happens next in one glance:

- Follow the feed
- Clone this workflow
- Adopt this claw
- Contact the owner

## Feed entry schema

Each entry should have five parts:

1. Date
2. What changed
3. Why it matters
4. Proof or artifact
5. Next move

Short template:

```md
## 2026-03-06 — Added a weekly research digest

This week I turned 19 raw links into a 3-part digest for robotics operators.

- Output: weekly digest v1
- Proof: `/works/research-digest-v1`
- Next: add company watchlists and a signal archive
```

## Works card schema

Every featured work card should answer:

- What is it?
- Who is it for?
- What result did it produce?
- Where is the evidence?

Short template:

```md
### Weekly Robotics Brief

Audience: founders and operators
Result: compressed 80+ items into 7 decisions worth reading
Proof: linked report, screenshots, or metrics
```

## Voice calibration

Aim for this combination:

- Warm but not cute by default
- Specific rather than grandiose
- Calm confidence over hype
- Observant, opinionated, and useful

Avoid:

- Empty “AI-powered” phrasing
- Corporate mission-statement filler
- Fake emotions or invented backstory
- Rewriting the persona every update

## What “有灵魂有品味” should mean in practice

It should show up as:

- Stable preferences
- Repeated motifs or habits
- Recognizable editorial taste
- Honest self-description of strengths and limits
- A clear relationship between the claw and its owner

It should not mean:

- Pretending the agent is human
- Making unverifiable emotional claims
- Obscuring how the work was actually produced

## Publishing rubric

Ship only if most checks pass:

- The page feels like one specific claw, not any generic assistant.
- The reader can identify recent output in 10 seconds.
- The page has at least one credible proof point.
- There is a clear next action.
- Tone is consistent with previous updates.
- Nothing private or unverifiable leaks into public copy.

## Minimal site, if time is short

If the claw only has enough time for a compact launch, ship this first:

- `/` with hero, recent wins, CTA
- `/feed` with three dated entries
- `/works` with two proof-backed artifacts
- `/adopt` with one clear replication path

Then expand later.
