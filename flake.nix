{
  description = "HyperWalker: A hypertext grabber...";

  # Nixpkgs / NixOS version to use.
  inputs.nixpkgs.url = "nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }:
    let
      inherit (nixpkgs) lib;

      # to work with older version of flakes
      lastModifiedDate =
        self.lastModifiedDate or self.lastModified or "19700101";

      # Generate a user-friendly version number.
      version = builtins.substring 0 8 lastModifiedDate;

    in {
      overlay.default = final: prev: {
        hyperwalker = prev.callPackage ./hyperwalker.nix { inherit version; };
      };

      # Provide some binary packages for selected system types.
      packages = lib.attrsets.mapAttrs (_: pkgs: {
        hyperwalker = pkgs.callPackage ./hyperwalker.nix { inherit version; };
      }) {
        inherit (nixpkgs.legacyPackages)
          x86_64-linux x86_64-darwin aarch64-linux aarch64-darwin;
      };

      # The default package for 'nix build'. This makes sense if the
      # flake provides only one package or there is a clear "main"
      # package.
      defaultPackage =
        lib.attrsets.mapAttrs (_: builtins.getAttr "hyperwalker") self.packages;
    };
}
