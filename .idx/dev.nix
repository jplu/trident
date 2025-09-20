{ pkgs, lib, ... }: {
  channel = "stable-25.05";

  packages = [
    pkgs.go
  ];

  imports = lib.optionals (builtins.pathExists ./dev.local.nix ) [ ./dev.local.nix ];

  env = {

  };

  idx = {
    extensions = [
      "golang.go"
    ];
  };
}
