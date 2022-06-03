{ lib, fetchurl, buildGoModule, nix-gitignore, makeWrapper, firefox, version ? "unstable"
}:

let freezeDry = fetchurl { url = "https://git.callpipe.com/dvn/hyperwalker/-/jobs/16551/artifacts/raw/build/freeze-dry/freeze-dry.umd.js"; hash = "sha256-Vc1vLLllFwekmDcyeB+PF9mYTvJFd3WLBAERORv9ER8="; };

in buildGoModule rec {
  pname = "HyperWalker";
  inherit version;

  # In 'nix develop', we don't need a copy of the source tree
  # in the Nix store.
  src = nix-gitignore.gitignoreSource ''
    /*.nix
    /flake.lock
  '' ./.;

  # This hash locks the dependencies of this package. It is
  # necessary because of how Go requires network access to resolve
  # VCS.  See https://www.tweag.io/blog/2021-03-04-gomod2nix/ for
  # details. Normally one can build with a fake sha256 and rely on native Go
  # mechanisms to tell you what the hash should be or determine what
  # it should be "out-of-band" with other tooling (eg. gomod2nix).
  # To begin with it is recommended to set this, but one must
  # remeber to bump this hash when your dependencies change.
  #vendorSha256 = pkgs.lib.fakeSha256;

  vendorSha256 = "sha256-yWl8xHSU480B9nkMGaFrJ0L+MmgmM7sLIvluub9S02c=";

  prePatch = ''
    mkdir -p js/dist
    cp ${freezeDry} js/dist/freeze-dry.umd.js
  '';

  nativeBuildInputs = [ makeWrapper ];
  postInstall = ''
    wrapProgram $out/bin/${pname} \
      --prefix PATH : ${lib.makeBinPath [ firefox ]}
  '';
}
