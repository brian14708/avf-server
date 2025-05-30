{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
    crane.url = "github:ipetkov/crane";
    rust-overlay = {
      url = "github:oxalica/rust-overlay";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    inputs@{
      flake-parts,
      crane,
      nixpkgs,
      rust-overlay,
      ...
    }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      imports = [
        inputs.treefmt-nix.flakeModule
      ];
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "aarch64-darwin"
        "x86_64-darwin"
      ];

      perSystem =
        {
          self',
          inputs',
          pkgs,
          system,
          ...
        }:
        let
          rustToolchain = (
            p:
            p.rust-bin.stable.latest.default.override {
              extensions = [ "rust-src" ];
            }
          );
          craneLib = (crane.mkLib pkgs).overrideToolchain rustToolchain;
          pname = "avf-server";
        in
        {
          _module.args.pkgs = import inputs.nixpkgs {
            inherit system;
            overlays = [ rust-overlay.overlays.default ];
          };

          devShells.default =
            (craneLib.devShell.override {
              mkShell = pkgs.mkShell.override {
                stdenv =
                  if pkgs.stdenv.hostPlatform.isDarwin then
                    pkgs.clangStdenv
                  else
                    pkgs.stdenvAdapters.useMoldLinker pkgs.clangStdenv;
              };
            })
              {
                packages = with pkgs; [
                  protobuf
                  pkg-config
                  ffmpeg-headless
                ];
                LIBCLANG_PATH = "${pkgs.llvmPackages.libclang.lib}/lib";
              };

          treefmt = {
            projectRootFile = "flake.nix";
            programs = {
              nixfmt.enable = true;
              rustfmt = {
                enable = true;
                package = rustToolchain pkgs;
              };
              buf.enable = true;
            };
          };

          apps = {
            default = {
              type = "app";
              program = "${self'.packages.default}/bin/${pname}";
              meta.description = "avf-server";
            };
          };

          packages = {
            default =
              let
                src =
                  with pkgs.lib.fileset;
                  toSource {
                    root = ./.;
                    fileset = unions [
                      (craneLib.fileset.commonCargoSources ./.)
                      ./proto
                    ];
                  };
                commonArgs = {
                  inherit pname src;
                  strictDeps = true;
                  nativeBuildInputs = with pkgs; [
                    protobuf
                    pkg-config
                  ];
                  stdenv = p: p.clangStdenv;
                  buildInputs = with pkgs; [ ffmpeg-headless ];
                  LIBCLANG_PATH = "${pkgs.llvmPackages.libclang.lib}/lib";
                };
              in
              craneLib.buildPackage (
                commonArgs
                // {
                  cargoArtifacts = craneLib.buildDepsOnly commonArgs;
                }
              );
            oci = pkgs.dockerTools.buildImage {
              name = pname;
              tag = "latest";
              copyToRoot = [ self'.packages.default ];
              config = {
                Cmd = [ "/bin/${pname}" ];
              };
            };
          };
          checks = {
          } // (pkgs.lib.mapAttrs' (n: pkgs.lib.nameValuePair "package-${n}") self'.packages);
        };
    };
}
