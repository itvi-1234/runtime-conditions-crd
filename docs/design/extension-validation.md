# Where does extension/Condition vocabulary get validated?

The spec's core only defines the top-level shape plus four core-reserved
Condition fields (`name`, `optional`, `kind`, `interface.type`). Everything
else about a Condition — the rest of `interface`, all of `configuration` —
is defined by whatever extensions the profile declares, each with its own
JSON Schema (2020-12) vocabulary.

That raises an obvious question when you turn this into a CRD: does the CRD
need a compiled-in validator for every extension?

Talked through three options with Colin:

1. Bake a validator per extension into the CR schema. No — the CRD would
   have to change every time any extension anywhere shipped, which kills
   the whole point of extensions being pluggable.
2. Validate the skeleton only, leave Condition vocabulary totally
   unchecked, let adapters catch problems later. Works, but a
   structurally-valid garbage Profile can sit in the cluster looking fine
   until some adapter finally trips over it — could be days later.
3. Validate the skeleton statically, add a first-party admission webhook
   that resolves `spec.extensions` and validates `spec.conditions` against
   their JSON Schema at admission time.

Went with option 3. It keeps the CRD schema stable (never needs to know
about extensions), and it's basically the same rule the spec already puts
on Adapters ("must resolve declared extensions before interpreting
extension-defined vocabulary") — we're just enforcing it one step earlier,
at admission instead of at adapter time.

## Fail closed

If the webhook can't fetch a declared extension's schema (host down,
network blip, whatever), it rejects the Profile rather than admitting it
with a warning. Confirmed with Colin over Slack.

This isn't free — fail-closed on a network-dependent lookup means an
extension host being briefly unreachable blocks every apply/update of every
Profile that depends on it, cluster-wide. So the webhook needs a caching
layer (resolved schemas, TTL'd, keyed by extension URL + version) from day
one, not as a later optimization.

## What the CRD already statically enforces vs. what the webhook still owns

The spec's own validation ordering (section 8) puts "is `kind` a non-empty
string", "is `interface.type` a non-empty string", "is `name` non-empty
when present" under *structural* validation — none of that needs an
extension resolved. So those four fields are typed directly on `Condition`
in the CRD, not left schemaless like I originally had them.

What's still schemaless, and still webhook work:
- anything inside `interface` beyond `type` (operations, engine, ...)
- all of `configuration`
- whether a `kind`/`interface.type` *value* is actually backed by a
  resolved extension
- condition name uniqueness across the array — skipped writing this as a
  CEL rule for now since a correct uniqueness check over an array with
  optional per-item names is easy to get subtly wrong with no cluster
  handy to test it against. Worth revisiting once there's a test cluster
  in CI; for now it's the webhook's job.

## Not built yet

The webhook itself — extension resolution, schema fetch/cache, actually
running JSON Schema 2020-12 validation, condition name uniqueness — doesn't
exist in this repo yet. This doc is just here so the CRD schema isn't
sitting on an undocumented assumption.
