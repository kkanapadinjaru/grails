/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,svelte}",
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        bg: {
          light: '#eff1f5',
          dark: '#24273a',
        },
        btn: {
          light: '#1e66f5',
          dark: '#8aadf4',
        },
        hi: {
          light: '#bdd6fb',
          dark: '#1d2d55',
        },
        sidebar: {
          light: '#e6e9ef',
          dark: '#1e2030',
        },
        text: {
          light: '#4c4f69',
          dark: '#cad3f5',
        },
        primary: {
          DEFAULT: '#3B82F6',
          hover: '#2563EB',
        },
        secondary: {
          DEFAULT: '#8B5CF6',
          hover: '#7C3AED',
        },
      },
    },
  },
  plugins: [
    require('@tailwindcss/typography'),
  ],
}
