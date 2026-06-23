{
  markdown =
    nixosConfigurations: featurePaths:
    let
      features = map (path: {
        name = builtins.head (builtins.match "(.+)\\.nix" (builtins.baseNameOf path));
        inherit path;
      }) featurePaths;
      configurations = builtins.attrNames nixosConfigurations;
      isImported =
        configuration: feature:
        builtins.elem (toString feature.path) (
          map toString nixosConfigurations.${configuration}.featureImports
        );
      row =
        configuration:
        "| ${nixosConfigurations.${configuration}.architecture} | ${configuration} |"
        + builtins.concatStringsSep "" (
          map (feature: " ${if isImported configuration feature then "✅" else ""} |") features
        );
    in
    builtins.concatStringsSep "\n" (
      [
        (
          "| Architecture | Configuration |"
          + builtins.concatStringsSep "" (map (feature: " ${feature.name} |") features)
        )
        ("| --- | --- |" + builtins.concatStringsSep "" (map (_: " --- |") features))
      ]
      ++ map row configurations
    )
    + "\n";
}
