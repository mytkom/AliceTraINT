/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./web/templates/**/*.{html,js}"],
  theme: {
    extend: {},
  },
  plugins: [
    require('@tailwindcss/forms'),
    require('@tailwindcss/typography'),
  ],
  safelist: [
    'bg-emerald-600',
    'bg-yellow-200',
    'bg-yellow-600',
    'bg-green-600',
    'bg-green-400',
    'bg-red-400',
    'bg-gray-400',
  ],
}

