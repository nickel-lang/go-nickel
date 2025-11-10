{
  description = "Go bindings to Nickel";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
    nickel.url = "github:tweag/nickel";
  };

  # use cached nickel
  nixConfig = {
    extra-substituters = [ "https://tweag-nickel.cachix.org" ];
    extra-trusted-public-keys = [ "tweag-nickel.cachix.org-1:GIthuiK4LRgnW64ALYEoioVUQBWs0jexyoYVeLDBwRA=" ];
  };

  outputs = inputs:
    let
      SYSTEMS = [
        "aarch64-darwin"
        "aarch64-linux"
        "x86_64-darwin"
        "x86_64-linux"
      ];

      lib = inputs.nixpkgs.lib;
      foreach = xs: f: with lib; foldr recursiveUpdate { } (map f xs);
      forSystems = systems: f: foreach systems (system: f system inputs.nixpkgs.legacyPackages.${system});
    in
    forSystems SYSTEMS (system: pkgs:
      {
        devShells.${system}.default = pkgs.mkShell {
          packages = with pkgs; [
            go
            gopls
          ];
          NICKEL_LIB=inputs.nickel.packages.${system}.nickel-lang-c;
        };
      });
}
