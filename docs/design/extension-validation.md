# Where does extension/Condition vocabulary get validated?

The spec's core only defines the top-level shape plus four core-reserved
Condition fields (`name`, `optional`, `kind`, `interface.type`). Everything
else about a Condition's `interface` is defined by whatever extensions the
profile declares, each with its own JSON Schema (2020-12) vocabulary.

That raises a question when turning this into a CRD: does the CRD need a
compiled-in validator for every extension?

Three options were considered:

1. Bake a validator per extension into the CR schema. Rejected - the CRD
   would have to change every time any extension anywhere shipped, which
   defeats the point of extensions being pluggable.
2. Validate the skeleton only, leave Condition vocabulary unchecked, let
   downstream consumers catch problems later. Works, but a
   structurally-valid Profile with garbage vocabulary can sit in the
   cluster looking fine until something finally trips over it.
3. Validate the skeleton statically, and have some component resolve the
   declared extensions and validate `conditions` against their JSON Schema
   before/at the point the Profile is admitted.

Went with the shape of option 3. It keeps the CRD schema stable (it never
needs to know about any specific extension), and it's the same rule the
spec already puts on Adapters: resolve declared extensions before
interpreting extension-defined vocabulary. Whether that validation runs as
an admission webhook in this repo, a separate service, or something else
entirely is still open - see below.

## What the CRD statically enforces vs. what's still open

The spec's validation ordering (section 8) classifies checking that `kind`
is a non-empty string, `interface.type` is a non-empty string, and `name`
(when present) is non-empty as *structural* validation - none of it needs
an extension resolved. So those fields are typed directly on `Condition` in
this CRD's schema.

Still not handled by this schema, and not yet built anywhere:
- whether a `kind`/`interface.type` *value* is actually backed by a
  resolved extension
- anything inside `interface` beyond `type` (extension-defined fields like
  operations, engine, etc.) - preserved via
  `x-kubernetes-preserve-unknown-fields` rather than pruned, but unvalidated
- condition name uniqueness across the array
- full JSON Schema semantic validation against declared extensions

## Open question: where does that validation live?

Whether this repo hosts a validating admission webhook, or whether that's a
separate project/repo entirely, hasn't been decided. This needs discussion
before any webhook code goes into this repo.
