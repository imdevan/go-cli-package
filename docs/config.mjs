const stage = process.env.NODE_ENV || "dev"
const isProduction = stage === "production"

export default {
  url: isProduction ? "https://devan.gg" : "http://localhost:4321",
  basePath: isProduction ? "/go-cli-package" : "/",
  github: "https://github.com/imdevan/go-cli-package/",
  githubDocs: "https://github.com/imdevan/go-cli-package/",
  title: "go-cli-package",
  description: "A go cli to publish your go cli package.",
}
