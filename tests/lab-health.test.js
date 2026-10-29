"use strict";

var assert = require("assert");
var contract = require("../lab-health.js");

function snapshot(overrides) {
  return Object.assign({
    schema_version:1,
    generated_at:"2026-07-24T20:00:00Z",
    nodes:[{
      name:"build-node",
      online:true,
      reachable:true,
      role:"worker",
      load:"low",
      health:"healthy",
      unexpected_private_field:"must not pass through"
    }]
  }, overrides || {});
}

var current = contract.parse(snapshot());
assert.equal(current.state, "snapshot");
assert.deepEqual(Object.keys(current.nodes[0]).sort(),
  ["health","load","name","online","reachable","role"]);

assert.equal(contract.parse(snapshot({schema_version:2})).state, "unsupported");
assert.equal(contract.parse(snapshot({generated_at:"not-a-date"})).state, "malformed");
assert.equal(contract.parse(snapshot({nodes:[{name:"partial"}]})).state, "malformed");
// Old snapshots stay visible: there is no automated publisher to keep them fresh.
assert.equal(contract.parse(snapshot({generated_at:"2020-01-01T00:00:00Z"})).state, "snapshot");

console.log("lab-health contract tests passed");
