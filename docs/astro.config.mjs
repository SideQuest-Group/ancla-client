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
          label: "Guides",
          items: [
            { label: "Getting Started", slug: "guides/getting-started" },
            { label: "Authentication", slug: "guides/authentication" },
            { label: "Configuration", slug: "guides/configuration" },
            { label: "Shell Completion", slug: "guides/shell-completion" },
          ],
        },
        {
          label: "CLI Reference",
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
              label: "Settings",
              collapsed: true,
              autogenerate: { directory: "cli/settings" },
            },
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
            { label: "Organizations", slug: "api/organizations" },
            { label: "Projects", slug: "api/projects" },
            { label: "Applications", slug: "api/applications" },
            { label: "Images", slug: "api/images" },
            { label: "Releases", slug: "api/releases" },
            { label: "Deployments", slug: "api/deployments" },
            { label: "Configuration", slug: "api/configuration" },
            { label: "Integrations", slug: "api/integrations" },
          ],
        },
        {
          label: "Python SDK",
          items: [
            { label: "Overview", slug: "sdk" },
          ],
        },
        {
          label: "OpenTofu Provider",
          items: [
            { label: "Overview", slug: "opentofu" },
          ],
        },
      ],
    }),
  ],
});
