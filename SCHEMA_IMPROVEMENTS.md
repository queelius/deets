# Schema Improvements: Identity, Cross-Referencing, and SEO

## Problem Statement

The current `me.toml` schema has grown organically and has several structural issues that make it harder for consuming tools (Claude Code, pyproject.toml generators, CITATION.cff writers, Hugo config, etc.) to reliably extract the right information.

### Current Pain Points

**1. Emails are scattered with ad-hoc naming**

```toml
[contact]
email = "lex@metafunctor.com"        # primary? personal? professional?
email_academic = "atowell@siue.edu"
email_gmail = "queelius@gmail.com"

[packages]
pypi_email = "lex@metafunctor.com"   # duplicates contact.email
```

A tool wanting "all emails" has to know about `email`, `email_academic`, `email_gmail`, and also `packages.pypi_email`. There's no way to query "give me all emails" or "give me the email tagged for PyPI use."

**2. Names/aliases lack structure for consuming tools**

```toml
[identity]
name = "Alexander Towell"
aka = ["Alex Towell"]
```

When populating `pyproject.toml` authors, CITATION.cff, or git config, tools need to know:
- Which name+email pairs go together?
- Which is the formal/legal name vs the common alias?
- Should both appear as separate authors, or is one preferred?

Currently `packages.pypi_author = "Alex Towell"` duplicates `identity.aka[0]` with no formal link.

**3. Missing descriptions on many fields**

The schema shows empty descriptions for `email_academic`, `email_gmail`, `pypi_author`, `pypi_email`, `scholar`, `zenodo`, `blog_alt`, `pypi`, `r_universe`. Tools using `deets describe` get nothing useful.

**4. No cross-referencing structure**

There's no way to express "queelius@gmail.com is the email I use for GitHub and PyPI" or "Alexander Towell is my ORCID name, Alex Towell is my GitHub name." The `packages` category is a workaround for this.

---

## Proposed Improvements

### Option A: Structured emails array (recommended)

Replace scattered email fields with a typed array:

```toml
[identity]
name = "Alexander Towell"
aka = ["Alex Towell"]

[[contact.emails]]
address = "lex@metafunctor.com"
label = "primary"
use = ["professional", "pypi", "cran"]

[[contact.emails]]
address = "queelius@gmail.com"
label = "gmail"
use = ["github", "personal"]

[[contact.emails]]
address = "atowell@siue.edu"
label = "academic"
use = ["academic", "orcid"]
```

Benefits:
- `deets get contact.emails` returns all emails
- `deets get contact.emails --filter use=pypi` returns the right one
- Each email has a clear purpose via `use` tags
- No more `packages.pypi_email` duplication

### Option B: Keep flat, add conventions

Keep the flat key-value model but add conventions:

```toml
[contact]
email = "lex@metafunctor.com"               # primary
emails = ["lex@metafunctor.com", "queelius@gmail.com", "atowell@siue.edu"]
email_labels = { "lex@metafunctor.com" = "primary", "queelius@gmail.com" = "github/personal", "atowell@siue.edu" = "academic" }
```

Simpler but less queryable.

### Option C: Platform-centric organization

Group by platform/context rather than data type:

```toml
[identity]
name = "Alexander Towell"
aka = ["Alex Towell"]

[platforms.github]
username = "queelius"
email = "queelius@gmail.com"
name = "Alex Towell"

[platforms.pypi]
username = "queelius"
email = "lex@metafunctor.com"
name = "Alex Towell"

[platforms.orcid]
id = "0000-0001-6443-9897"
email = "atowell@siue.edu"
name = "Alexander Towell"
```

Most explicit for cross-referencing, but duplicates names/emails.

---

## Consuming Tool Requirements

These are the actual queries tools need to answer:

| Use Case | Query | Current Answer |
|----------|-------|----------------|
| pyproject.toml authors | All name+email pairs for author list | Manual: combine identity.name + contact.email, identity.aka[0] + contact.email_gmail |
| CITATION.cff | Formal name + ORCID | identity.name + academic.orcid (works) |
| git config | Name + preferred email | identity.name + contact.email (works) |
| GitHub profile | Username + display name | web.github + identity.aka[0] (fragile) |
| "All my emails" | Every email address | Must know all `email*` keys across categories |
| "Email for PyPI" | Context-specific email | packages.pypi_email (duplicates contact.email) |
| Paper authorship | Formal name + institution + ORCID | Works, but scattered across 3 categories |
| SEO/discoverability | All public profile URLs | Must enumerate web.* manually |

---

## Specific Requests

1. **All fields should have descriptions** -- empty descriptions in schema output are unhelpful
2. **Emails should be queryable by purpose** -- "give me my PyPI email" without hardcoding `packages.pypi_email`
3. **Name+email pairing** -- tools should be able to get author tuples like `[("Alexander Towell", "lex@metafunctor.com"), ("Alex Towell", "queelius@gmail.com")]`
4. **Eliminate duplication** -- `packages.pypi_email` shouldn't duplicate `contact.email`; instead, tag the email with its use
5. **Profile URLs should be full URLs** -- some are handles (`queelius`), some are full URLs (`https://metafunctor.com`), some are mixed (`queelius.bsky.social`). Normalize to include a `url` field per platform
6. **Add missing identifiers** -- LinkedIn URL, personal website canonical URL, Semantic Scholar ID, dblp ID, arXiv author ID (if applicable)

---

## Priority

**High:** Fix email structure + add descriptions (directly impacts every `pyproject.toml` and `CITATION.cff` generation)

**Medium:** Name+email pairing for author lists, profile URL normalization

**Low:** Platform-centric reorganization, additional identifiers
