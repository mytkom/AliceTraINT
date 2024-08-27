{
  description = "A Go full-stack web app with PostgreSQL, HTMX, and Nix flake support";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
    let
      pkgs = import nixpkgs { inherit system; };
    in
    {
      devShell = pkgs.mkShell {
        buildInputs = [
          pkgs.go_1_22
          pkgs.docker
          pkgs.postgresql
          pkgs.migrate
          pkgs.golangci-lint
          pkgs.sqlc
          pkgs.makeWrapper # To help with scripts or any wrapping needs
        ];

        shellHook = ''
          export PATH=$PWD/bin:$PATH
        '';
      };

      packages.default = pkgs.buildGoModule {
        pname = "alice-traint";
        version = "0.1.0";
        src = ./.;
        vendorSha256 = null;

        goModVendor = true;

        buildPhase = ''
          go build -o $out/bin/alice-traint ./cmd/alice-traint
        '';

        meta = with pkgs.lib; {
          description = "A Go full-stack web app";
          license = licenses.mit;
          maintainers = [ maintainers.mytkom ];
        };
      };
    });
}

