// Mirrors fixtures/demo-project: what its `monom complete` prints, and what
// each script echoes when run.
export const commands: Record<string, string> = {
  "db/migrate": "running db migrations...",
  "db/seed": "seeding database...",
  "infra/cloud/deploy": "deploying to cloud...",
  "infra/cloud/teardown": "tearing down cloud infrastructure...",
  "infra/local/start": "starting local environment...",
  "infra/local/stop": "stopping local environment...",
  release: "releasing...",
};

export const paths = Object.keys(commands);
