{
  description = "A very basic flake";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:

    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};

        go_1_23 = pkgs.go_1_22.overrideAttrs (old: rec {
          src = pkgs.fetchurl {
            url = "https://go.dev/dl/go${version}.src.tar.gz";
            hash = "sha256-9pnOJWD8Iq2CwGseBLYxi4Xn9obLy0/OFWWCEyxX2Ps=";
          };
          version = "1.23rc2";
          doCheck = false;
        });

        gopls_1_23 = pkgs.gopls.override { buildGoModule = buildGo123Module; };
        buildGo123Module = pkgs.buildGo122Module.override { go = go_1_23; };
      in
      {
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            self.formatter.${system}
            go_1_23
            gopls_1_23
            gotools
          ];
        };

        formatter = pkgs.nixfmt-rfc-style;
      }
    );
}
