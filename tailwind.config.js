/** @type {import('tailwindcss').Config} */
// copied from https://github.com/Phillip-England/templ-quickstart/blob/main/tailwind.config.js
module.exports = {
    content: [
        "./internal/**/*.{go,js,templ,html}"
    ],
    theme: {
      extend: {
        colors: {
            bbgreen: "#bef264",
            bbgreenoff: "#22c55e",
            bbyellow: "#fef08a",
            bbyellowoff: "#fb923c",
            bbpink: "#fda4af",
            bbpinkoff: "#f43f5e"
        }
      },
    },
    plugins: [],
  }