import { defineConfig } from "astro/config";
import starlight from "@astrojs/starlight";

export default defineConfig({
  site: "https://docs.ancla.dev",
  legacy: { collections: true },
  integrations: [
    starlight({
      title: "Ancla",
      favicon: "/favicon.svg",
      customCss: ["./src/styles/custom.css"],
      head: [
        {
          tag: "meta",
          attrs: { name: "theme-color", content: "#0b1120" },
        },
      ],
      social: [
        {
          icon: "github",
          label: "GitHub",
          href: "https://github.com/SideQuest-Group/ancla-client",
        },
      ],
      sidebar: [
        {
          label: "Platform",
          items: [
            { label: "Overview", slug: "platform" },
            { label: "Deploy Pipeline", slug: "platform/deploy-pipeline" },
            { label: "Secrets & Config", slug: "platform/secrets-and-config" },
          ],
        },
        {
          label: "Guides",
          items: [
            { label: "Getting Started", slug: "guides/getting-started" },
            { label: "Authentication", slug: "guides/authentication" },
            { label: "Configuration", slug: "guides/configuration" },
            { label: "Project Linking", slug: "guides/project-linking" },
            { label: "Development Workflow", slug: "guides/dev-workflow" },
            { label: "Remote Access", slug: "guides/remote-access" },
            { label: "Scripting & Automation", slug: "guides/scripting" },
            { label: "Shell Completion", slug: "guides/shell-completion" },
          ],
        },
        {
          label: "CLI Reference",
          collapsed: true,
          items: [
            { label: "Overview", slug: "cli/ancla" },
            {
              label: "Apps",
              collapsed: true,
              autogenerate: { directory: "cli/apps" },
            },
            {
              label: "Config",
              collapsed: true,
              autogenerate: { directory: "cli/config" },
            },
            {
              label: "Deployments",
              collapsed: true,
              autogenerate: { directory: "cli/deployments" },
            },
            {
              label: "Images",
              collapsed: true,
              autogenerate: { directory: "cli/images" },
            },
            {
              label: "Orgs",
              collapsed: true,
              autogenerate: { directory: "cli/orgs" },
            },
            {
              label: "Projects",
              collapsed: true,
              autogenerate: { directory: "cli/projects" },
            },
            {
              label: "Releases",
              collapsed: true,
              autogenerate: { directory: "cli/releases" },
            },
            {
              label: "Cache",
              collapsed: true,
              autogenerate: { directory: "cli/cache" },
            },
            {
              label: "Settings",
              collapsed: true,
              autogenerate: { directory: "cli/settings" },
            },
            { label: "init", slug: "cli/ancla_init" },
            { label: "link", slug: "cli/ancla_link" },
            { label: "unlink", slug: "cli/ancla_unlink" },
            { label: "status", slug: "cli/ancla_status" },
            { label: "logs", slug: "cli/ancla_logs" },
            { label: "run", slug: "cli/ancla_run" },
            { label: "open", slug: "cli/ancla_open" },
            { label: "docs", slug: "cli/ancla_docs" },
            { label: "down", slug: "cli/ancla_down" },
            { label: "list", slug: "cli/ancla_list" },
            { label: "ssh", slug: "cli/ancla_ssh" },
            { label: "shell", slug: "cli/ancla_shell" },
            { label: "dbshell", slug: "cli/ancla_dbshell" },
            { label: "login", slug: "cli/ancla_login" },
            { label: "whoami", slug: "cli/ancla_whoami" },
            { label: "completion", slug: "cli/ancla_completion" },
            { label: "version", slug: "cli/ancla_version" },
          ],
        },
        {
          label: "API Reference",
          collapsed: true,
          items: [
            { label: "Overview", slug: "api" },
            { label: "Authentication", slug: "api/authentication" },
            { label: "Workspaces", slug: "api/workspaces" },
            { label: "Teams", slug: "api/teams" },
            { label: "Projects", slug: "api/projects" },
            { label: "Environments", slug: "api/environments" },
            { label: "Services", slug: "api/services" },
            { label: "Pipeline", slug: "api/pipeline" },
            { label: "Promotions", slug: "api/promotions" },
            { label: "Observability", slug: "api/observability" },
            { label: "Integrations", slug: "api/integrations" },
          ],
        },
        {
          label: "SDKs",
          items: [
            { label: "Overview", slug: "sdk" },
            { label: "Python", slug: "sdk/python" },
            { label: "Go", slug: "sdk/go" },
            { label: "TypeScript", slug: "sdk/typescript" },
          ],
        },
        {
          label: "Terraform Provider",
          items: [
            { label: "Overview", slug: "opentofu" },
          ],
        },
        {
          label: "Competitive Landscape",
          slug: "competitive-landscape",
        },
      ],
    }),
  ],
});
