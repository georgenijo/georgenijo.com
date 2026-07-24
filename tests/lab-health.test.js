"use strict";

var assert = require("assert");
var contract = require("../lab-health.js");
var now = Date.parse("2026-07-24T20:10:00Z");

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

var current = contract.parse(snapshot(), now);
assert.equal(current.state, "current");
assert.deepEqual(Object.keys(current.nodes[0]).sort(),
  ["health","load","name","online","reachable","role"]);

assert.equal(contract.parse(snapshot({schema_version:2}), now).state, "unsupported");
assert.equal(contract.parse(snapshot({generated_at:"not-a-date"}), now).state, "malformed");
assert.equal(contract.parse(snapshot({nodes:[{name:"partial"}]}), now).state, "malformed");
assert.equal(contract.parse(snapshot(), Date.parse("2026-07-24T20:16:00Z")).state, "stale");
assert.equal(contract.parse(snapshot(), Date.parse("2026-07-24T19:58:00Z")).state, "stale");

console.log("lab-health contract tests passed");
