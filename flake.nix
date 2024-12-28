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

      alienpyPkg = pkgs.python312Packages.buildPythonPackage rec {
        pname = "alienpy";
        version = "1.6.1"; # Replace with the version you need

        src = pkgs.fetchPypi {
          inherit pname version;
          sha256 = "yRf9qCEpzwGgmis4kUHzqbdTzqbcCwbuwoEd7xcyOUQ="; # Replace with the actual hash
        };

        propagatedBuildInputs = with pkgs.python312Packages; [
          requests
          websockets
          rich
          async_stagger
          cryptography
          pyopenssl
          xrootd
        ];

        meta = with pkgs.lib; {
          description = "AlienPy is a Python package to interact with the ALICE Grid.";
          license = licenses.mit;
          homepage = "https://pypi.org/project/alienpy/";
        };
      };
    in
    {
      devShell = pkgs.mkShell {
        hardeningDisable = [ "fortify" ];
        buildInputs = with pkgs; [
          go_1_22
          gcc
          docker
          postgresql
          migrate
          golangci-lint
          sqlc
          tailwindcss
          alienpyPkg
          makeWrapper
          cypress
        ];

        shellHook = ''
          export PATH=$PWD/bin:$PATH
          alien_ls /
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

