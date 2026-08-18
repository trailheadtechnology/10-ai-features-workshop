// Demo starting point: the "you might also like" box, picking trails at random.
// Run: npm run starter [-- trail id or name]   (default: trail-0117, Avalanche Lake Trail)

import { readFileSync } from "node:fs";
import { resolve } from "node:path";

const DATA = resolve(import.meta.dirname, "../../data");
type Trail = { id: string; name: string; park: string; distance_mi: number; elevation_ft: number; difficulty: string; features: string[]; description: string };
const trails: Trail[] = JSON.parse(readFileSync(resolve(DATA, "trails.json"), "utf8"));

const query = process.argv.slice(2).join(" ") || "trail-0117";
const target = trails.find((t) => t.id.toLowerCase() === query.toLowerCase() || t.name.toLowerCase().includes(query.toLowerCase()));
if (!target) throw new Error(`No trail matches '${query}'.`);

console.log(`You liked: ${target.name} (${target.park})`);
console.log("You might also like (picked at random, which is the current feature):\n");

const others = trails.filter((t) => t.id !== target.id).sort(() => Math.random() - 0.5).slice(0, 5);
for (const t of others) console.log(`  ${t.name} (${t.park}, ${t.difficulty})`);
