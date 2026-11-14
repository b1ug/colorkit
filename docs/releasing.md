# Release sequence

1. Finish the formal library source, public API documentation, and regression
   tests. Compatibility prototypes are separate evidence.
2. Pass the source-bound local checks and CI: real Go 1.13, maintained Go,
   native platforms, and actual WASM execution. Confirm go.mod has no external
   dependencies and preserve the source/data hashes.
3. Select the first module version after acceptance. The proposed first version
   is `v0.1.0`; it is not an existing tag merely because this file mentions it.
4. Create the version tag from the accepted commit. Confirm the tagged source
   matches the verified commit and can be fetched by its intended consumers.
5. Update blink1-go to the colorkit tag in a separate dependency/integration PR;
   validate that combination before changing nb1.
6. Update nb1 to the validated SDK version and run its own software and bounded
   hardware acceptance. Treat SDK version tagging as a separate release step.

The repository starts private. Public downstream dependencies need a source
that their users can fetch without private credentials; repository publication
is separate from testing or creating a private version tag. Do not update
public downstream go.mod files to an inaccessible private module.

Do not move a published tag. New fixes require another version.
