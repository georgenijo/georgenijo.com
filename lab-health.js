(function(root, factory) {
  var api = factory();
  if (typeof module === "object" && module.exports) module.exports = api;
  else root.GeorgeLabHealth = api;
})(typeof globalThis !== "undefined" ? globalThis : this, function() {
  "use strict";

  var LOADS = {low:true, moderate:true, high:true, unknown:true};
  var HEALTH = {healthy:true, degraded:true, unknown:true, unreachable:true};

  function text(value, field) {
    if (typeof value !== "string" || value.length < 1 || value.length > 80) {
      throw new Error("invalid " + field);
    }
    return value;
  }

  function parse(snapshot) {
    if (!snapshot || typeof snapshot !== "object" || Array.isArray(snapshot)) {
      return {state:"malformed", nodes:[]};
    }
    if (snapshot.schema_version !== 1) {
      return {state:"unsupported", nodes:[]};
    }
    var generatedMs = Date.parse(snapshot.generated_at);
    if (!Number.isFinite(generatedMs) || !Array.isArray(snapshot.nodes)) {
      return {state:"malformed", nodes:[]};
    }
    try {
      var nodes = snapshot.nodes.map(function(node) {
        if (!node || typeof node !== "object" || Array.isArray(node) ||
            typeof node.online !== "boolean" || typeof node.reachable !== "boolean" ||
            !LOADS[node.load] || !HEALTH[node.health]) {
          throw new Error("invalid node");
        }
        // Copy only the public contract allowlist. Unexpected fields never reach the UI.
        return {
          name:text(node.name, "name"),
          role:text(node.role, "role"),
          online:node.online,
          reachable:node.reachable,
          load:node.load,
          health:node.health
        };
      });
      // Snapshots are published by hand, so there is no freshness window;
      // the UI labels the data with its collection date instead.
      return {
        state:"snapshot",
        generatedAt:snapshot.generated_at,
        nodes:nodes
      };
    } catch (_) {
      return {state:"malformed", nodes:[]};
    }
  }

  return {parse:parse};
});
