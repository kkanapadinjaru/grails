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
          light: 'var(--color-bg-light)',
          dark: 'var(--color-bg-dark)',
        },
        btn: {
          light: 'var(--color-btn-light)',
          dark: 'var(--color-btn-dark)',
        },
        hi: {
          light: 'var(--color-hi-light)',
          dark: 'var(--color-hi-dark)',
        },
        sidebar: {
          light: 'var(--color-sidebar-light)',
          dark: 'var(--color-sidebar-dark)',
        },
        text: {
          light: 'var(--color-text-light)',
          dark: 'var(--color-text-dark)',
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
