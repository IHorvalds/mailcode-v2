let
  pkgs = import <nixpkgs> { config = {}; overlays = []; };
in
  pkgs.mkShellNoCC {
    packages = with pkgs; [
      nodejs_23
      go
      gnumake
    ];
  }
