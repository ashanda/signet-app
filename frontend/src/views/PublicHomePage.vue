<script setup>
// Ports welcome.blade.php — the public landing page, a standalone design
// with its own stylesheet (resources/site/css/style.css), fonts and scripts,
// sharing nothing with the app skin (main.js disables the app skin on this
// route via meta.skin 'site').
//
// The markup is ./public/welcome.html, generated verbatim from the Blade
// view's <body>; the behaviour is ./public/siteScripts.js, generated from
// resources/site/js/main.js + live-prices.js (see the headers of those
// files). The stylesheet is attached while this page is mounted and the
// markup is only rendered once it has loaded, so the canvases and scroll
// triggers measure the styled layout, as they do on a normal page load.
import { ref, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { gsap } from 'gsap'
import { ScrollTrigger } from 'gsap/ScrollTrigger'
import welcomeHtml from './public/welcome.html?raw'
import { initSite } from './public/siteScripts'

const router = useRouter()
const html = welcomeHtml.replace('__YEAR__', String(new Date().getFullYear()))
const ready = ref(false)
const root = ref(null)

const STYLESHEET = '/resources/site/css/style.css'
const TITLE = 'Signet — Digital-Asset Intelligence Platform'
const DESCRIPTION =
  'Signet is an institutional-grade digital-asset platform — market intelligence, portfolio tooling and a curated cryptocurrency ecosystem in one professional workspace.'

let link = null
let meta = null
let previousTitle = ''
let cleanup = null
let unmounted = false

function loadStylesheet() {
  return new Promise((resolve) => {
    link = document.createElement('link')
    link.rel = 'stylesheet'
    link.href = STYLESHEET
    link.onload = resolve
    link.onerror = resolve
    document.head.appendChild(link)
  })
}

// Links to app pages (/login, /register) navigate inside the SPA; in-page
// anchors (#about, ...) keep the browser's native jump.
function onClick(event) {
  const a = event.target.closest && event.target.closest('a[href^="/"]')
  if (!a || event.defaultPrevented || event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || a.target) return
  event.preventDefault()
  router.push(a.getAttribute('href'))
}

onMounted(async () => {
  previousTitle = document.title
  document.title = TITLE
  meta = document.createElement('meta')
  meta.name = 'description'
  meta.content = DESCRIPTION
  document.head.appendChild(meta)

  await loadStylesheet()
  if (unmounted) return
  ready.value = true
  await nextTick()
  if (unmounted) return
  gsap.registerPlugin(ScrollTrigger)
  cleanup = initSite({ gsap, ScrollTrigger })
})

onBeforeUnmount(() => {
  unmounted = true
  if (cleanup) cleanup()
  if (link) link.remove()
  if (meta) meta.remove()
  document.title = previousTitle || 'signetint'
})
</script>

<template>
  <!-- eslint-disable-next-line vue/no-v-html -->
  <div v-if="ready" ref="root" class="signet-site" @click="onClick" v-html="html"></div>
</template>
