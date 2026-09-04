<script setup>
// Ports welcome.blade.php + layouts/main.blade.php (route `guest`) — the
// public marketing landing page, which is a completely separate "skin" from
// the rest of the app (its own Bootstrap build, fonts, Swiper carousels,
// jQuery + slicknav mobile menu — see ui_spec.md "Two separate front-end
// skins"). Loaded/unloaded at runtime (not via the global main.js imports
// that every other page uses) so its vendor CSS/JS never bleeds into the
// Volt-themed dashboard/auth pages, and vice versa.
//
// Sections that are HTML-commented-out in the original (Exclusive Partners,
// Our Services, Download Apps, Crypto Calculator, Our Team, Testimonials,
// Latest Posts) are skipped entirely, matching ui_spec.md.
//
// Deliberately NOT reproduced (transient/decorative effects whose final
// rendered state is identical with or without them — see conversation):
// the magic-cursor follower dot (gsap), the text-anime letter-cascade
// reveal (splitType + ScrollTrigger), hero parallax scroll offset, and the
// newsletter/video-popup/counter scripts (their sections are static or
// non-functional `action="#"` in the original anyway). WOW.js scroll-reveal
// and the two price-carousel Swiper instances ARE wired up for real since
// those affect layout/visibility, not just a transient animation.
//
// .reveal (the about-us image wrapper) is the one exception that is NOT
// safe to skip outright: custom.css sets it permanently `visibility:hidden`
// and only function.js's GSAP+ScrollTrigger "Image Reveal Animation" block
// ever flips it back with `autoAlpha:1` — skip that script and the images
// stay invisible forever, unlike WOW.js's `.wow` (which self-resolves via
// real scroll events). Rather than pull in gsap+ScrollTrigger for one
// slide-in transient, the <style> block below forces the end state
// (visible, no slide-in) directly.
import { onMounted, onUnmounted } from 'vue'

const ASSET_BASE = '/resources/main'
const injected = []

function addLink(href, rel = 'stylesheet', crossorigin = false) {
  const link = document.createElement('link')
  link.rel = rel
  link.href = href
  if (crossorigin) link.crossOrigin = 'anonymous'
  link.dataset.signetPublicSkin = 'true'
  document.head.appendChild(link)
  injected.push(link)
}

function addScript(src) {
  return new Promise((resolve, reject) => {
    const script = document.createElement('script')
    script.src = src
    script.dataset.signetPublicSkin = 'true'
    script.onload = resolve
    script.onerror = reject
    document.head.appendChild(script)
    injected.push(script)
  })
}

let swiperInstances = []

onMounted(async () => {
  // Google Fonts (layouts/main.blade.php's preconnect + DM Sans stylesheet)
  // — custom.css's --default-font is "DM Sans", sans-serif throughout, so
  // without this the whole page silently falls back to the browser default.
  addLink('https://fonts.googleapis.com/', 'preconnect')
  addLink('https://fonts.gstatic.com/', 'preconnect', true)
  addLink("https://fonts.googleapis.com/css2?family=DM+Sans:opsz,wght@9..40,100..1000&display=swap")

  // CSS (appended after the global volt.css/custome.css imports, so its
  // shared Bootstrap base classes win the cascade only while this page is
  // mounted).
  ;[
    'css/bootstrap.min.css',
    'css/slicknav.min.css',
    'css/swiper-bundle.min.css',
    'css/all.min.css',
    'css/animate.css',
    'css/magnific-popup.css',
    'css/custom.css',
  ].forEach((f) => addLink(`${ASSET_BASE}/${f}`))

  // Preloader: the original fades this out via jQuery's $(window).on('load')
  // handler; there's no equivalent "page load" event in an SPA route
  // transition, so it's just faded out shortly after mount instead.
  requestAnimationFrame(() => {
    setTimeout(() => {
      const el = document.querySelector('.signet-public .preloader')
      if (el) {
        el.style.transition = 'opacity 0.6s'
        el.style.opacity = '0'
        setTimeout(() => el.remove(), 600)
      }
    }, 300)
  })

  try {
    await addScript(`${ASSET_BASE}/js/jquery-3.7.1.min.js`)
    await addScript(`${ASSET_BASE}/js/jquery.slicknav.js`)
    await addScript(`${ASSET_BASE}/js/swiper-bundle.min.js`)
    await addScript(`${ASSET_BASE}/js/wow.js`)

    // Slicknav mobile menu (function.js: $('#menu').slicknav({label:'', prependTo:'.responsive-menu'}))
    window.jQuery('#menu').slicknav({ label: '', prependTo: '.responsive-menu' })

    // WOW.js scroll-reveal for every .wow element (fadeInUp etc.)
    new window.WOW().init()

    // Price carousels — config ported verbatim from resources/main/js/function.js
    swiperInstances.push(
      new window.Swiper('.price-carousel.price-carousel-left .swiper', {
        slidesPerView: 1.5,
        centeredSlides: true,
        spaceBetween: 30,
        speed: 2500,
        loop: true,
        autoplay: { delay: 0 },
        allowTouchMove: false,
        disableOnInteraction: true,
        breakpoints: {
          768: { slidesPerView: 3 },
          991: { slidesPerView: 4 },
          1024: { slidesPerView: 5 },
          1440: { slidesPerView: 6 },
        },
      }),
    )
    swiperInstances.push(
      new window.Swiper('.price-carousel.price-carousel-right .swiper', {
        slidesPerView: 1.5,
        centeredSlides: true,
        speed: 2500,
        spaceBetween: 30,
        loop: true,
        autoplay: { delay: 0, reverseDirection: true },
        allowTouchMove: false,
        disableOnInteraction: true,
        breakpoints: {
          768: { slidesPerView: 3 },
          991: { slidesPerView: 4 },
          1024: { slidesPerView: 5 },
          1440: { slidesPerView: 6 },
        },
      }),
    )
  } catch {
    // Vendor script failed to load (offline/CDN-blocked) — page still
    // renders with static markup, just without the carousel/menu JS.
  }
})

onUnmounted(() => {
  swiperInstances.forEach((s) => s.destroy?.(true, true))
  swiperInstances = []
  injected.forEach((el) => el.remove())
  injected.length = 0
})

const tickerCoins = [
  'Bitcoin', 'Etherium', 'Tether', 'BNB', 'Solana', 'USD Coin',
  'Cardano', 'Cardano', 'Dogecoin', 'Tron', 'Polygon', 'Shiba INU',
  'Lite Coin', 'Stacks',
]

const howItWorks = [
  {
    items: [
      '$10 subscription → Earn 20% profit (basic signals provided).',
      '$100 subscription → Earn 40% profit (advanced signals provided).',
      '$1,000 subscription → Earn 60% profit (VIP signals provided).',
      '$5,000 subscription → Earn 70% profit (VVIP signals provided).',
      '$100,000 subscription → Earn up to 90% profit.',
    ],
    trailing: null,
  },
  {
    items: [
      'If you start with a $100 subscription, once you earn $400, you need to renew the $100 subscription.',
      'The remaining $300 stays in your account to grow your fund.',
    ],
    trailing: 'As a result, you will be able to grow your account balance to $5,000 or more.',
  },
  {
    items: [
      'The 1st person you introduce → You earn the commission',
      'The 2nd person → Commission goes to your introducer',
      'The 3rd person → You earn the commission',
      'The 4th person → Commission goes to your introducer',
    ],
    trailing: 'This pattern continues (2nd, 5th, 10th, 15th, etc., go to your introducer).',
  },
]

// Left/right price carousels, colors transcribed verbatim from welcome.blade.php
// (not derived — each item's green/red is hardcoded in the original markup).
const priceLeft = [
  { icon: 1, name: 'Bitcoin', color: 'green' },
  { icon: 2, name: 'Etherium', color: 'red' },
  { icon: 3, name: 'Tether', color: 'green' },
  { icon: 4, name: 'BNB', color: 'green' },
  { icon: 5, name: 'Solana', color: 'red' },
  { icon: 6, name: 'USD Coin', color: 'green' },
  { icon: 7, name: 'Cardano', color: 'green' },
  { icon: 8, name: 'Cardano', color: 'red' },
  { icon: 9, name: 'Dogecoin', color: 'green' },
  { icon: 10, name: 'Tron', color: 'green' },
  { icon: 11, name: 'Polygon', color: 'red' },
  { icon: 12, name: 'Shiba INU', color: 'green' },
]
const priceRight = [
  { icon: 13, name: 'Lite Coin', color: 'red' },
  { icon: 14, name: 'Stacks', color: 'green' },
  { icon: 15, name: 'Toncoin', color: 'red' },
  { icon: 16, name: 'Filecoin', color: 'green' },
  { icon: 17, name: 'Hedera', color: 'red' },
  { icon: 18, name: 'DigiByte', color: 'green' },
  { icon: 19, name: 'Centrifuge', color: 'red' },
  { icon: 20, name: 'Flux', color: 'green' },
  { icon: 21, name: 'Compound', color: 'red' },
  { icon: 22, name: 'Frax Share', color: 'green' },
  { icon: 23, name: 'Gemini Dollar', color: 'green' },
  { icon: 24, name: 'Maker', color: 'red' },
]
</script>

<template>
  <div class="signet-public">
    <div class="preloader">
      <div class="loading-container">
        <div class="loading"></div>
        <div id="loading-icon"><img src="/resources/main/images/loader.svg" alt="" /></div>
      </div>
    </div>

    <header class="main-header">
      <div class="header-sticky">
        <nav class="navbar navbar-expand-lg">
          <div class="container">
            <RouterLink class="navbar-brand" to="/">
              <img src="/resources/assets/img/brand/small-logo.png" alt="Logo" />
            </RouterLink>

            <div class="collapse navbar-collapse main-menu">
              <ul class="navbar-nav mr-auto" id="menu">
                <li class="nav-item"><RouterLink class="nav-link" to="/">Home</RouterLink></li>
                <li class="nav-item"><a class="nav-link" href="javascript:void(0);">About us</a></li>
                <li class="nav-item"><a class="nav-link" href="javascript:void(0);">Contact</a></li>
                <li class="nav-item"><RouterLink class="nav-link" to="/login">Login</RouterLink></li>
              </ul>
            </div>

            <div class="navbar-toggle"></div>
          </div>
        </nav>

        <div class="responsive-menu"></div>
      </div>
    </header>

    <!-- Hero Section -->
    <div class="hero parallaxie">
      <div class="container">
        <div class="row align-items-center">
          <div class="col-12">
            <div class="hero-content">
              <div class="section-title">
                <h3 class="wow fadeInUp">Welcome to Signet</h3>
                <h1 class="text-anime">Easy Way to <span>Signet </span> <br />Key To Success.</h1>
              </div>
              <div class="hero-content-body wow fadeInUp" data-wow-delay="0.5s">
                <p>Your journey to financial freedom starts with a single step&mdash;take action today and build the future you deserve!</p>
              </div>
              <div class="hero-content-footer wow fadeInUp" data-wow-delay="0.75s">
                <RouterLink to="/register" class="btn-default">Join for Free</RouterLink>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Coin Ticker -->
    <div class="coin-ticker">
      <div class="container-fluid">
        <div class="row no-gap">
          <div class="col-md-12">
            <div class="coin-ticker-box">
              <div class="scrolling-ticker-box">
                <div v-for="pass in 2" :key="pass" class="scrolling-content">
                  <div v-for="(coin, i) in tickerCoins" :key="`${pass}-${i}`" class="scrolling-item">
                    <div class="icon-box">
                      <img :src="`/resources/main/images/icon-ticker-${(i % 14) + 1}.svg`" alt="" />
                    </div>
                    <p>{{ coin }}</p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- About Signet -->
    <div class="about-us">
      <div class="container">
        <div class="row">
          <div class="col-md-12">
            <div class="section-title">
              <h3 class="wow fadeInUp">About Signet</h3>
              <h2 class="text-anime">Simple. Faster. Secure</h2>
            </div>
          </div>
        </div>

        <div class="row align-items-center">
          <div class="col-lg-6 col-12">
            <div class="about-images">
              <div class="about-image">
                <figure class="image-anime reveal">
                  <img src="/resources/main/images/about-us-1.jpg" alt="" />
                </figure>
              </div>
              <div class="about-image">
                <figure class="image-anime reveal">
                  <img src="/resources/main/images/about-us-2.jpg" alt="" />
                </figure>
              </div>
            </div>
          </div>

          <div class="col-lg-6 col-12">
            <div class="about-content">
              <div class="about-body wow fadeInUp" data-wow-delay="0.25s">
                <p>SIGNET provides highly valuable trading signals for just $10.Additionally, they claim to help you grow your account balance to $5,000.</p>
                <p>To achieve this, you don&rsquo;t need to leave your current job. You simply have to complete the tasks they assign, allowing you to:</p>
              </div>
              <div class="about-list-item wow fadeInUp" data-wow-delay="0.5s">
                <ul>
                  <li>Utilize their exclusive trading signals to earn profits.</li>
                  <li>Complete assigned tasks to grow your account balance to $5,000 or more.</li>
                </ul>
              </div>
              <div class="about-footer wow fadeInUp" data-wow-delay="0.75s">
                <a href="/resources/files/signet.pdf" class="btn-default" target="_blank">Read More</a>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- How It Works -->
    <div class="how-it-works">
      <div class="container">
        <div class="row">
          <div class="col-md-12">
            <div class="section-title">
              <h3 class="wow fadeInUp">How it Works</h3>
              <h2 class="text-anime">SIGNET Fund Building Plan</h2>
            </div>
          </div>
        </div>

        <div class="row">
          <div v-for="(step, i) in howItWorks" :key="i" class="col-md-4">
            <div class="how-it-works-item wow fadeInUp" data-wow-delay="0.25s">
              <div class="icon-box">
                <img src="/resources/main/images/icon-how-it-work-1.svg" alt="" />
              </div>
              <div class="about-list-item wow fadeInUp" data-wow-delay="0.5s">
                <ul>
                  <li v-for="(li, j) in step.items" :key="j">{{ li }}</li>
                </ul>
              </div>
              <p v-if="step.trailing">{{ step.trailing }}</p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Price Section -->
    <div class="price-section">
      <div class="container-fluid">
        <div class="row">
          <div class="col-md-12">
            <div class="section-title">
              <h3 class="wow fadeInUp">Price</h3>
              <h2 class="text-anime">Explore Cryptocurrency Price</h2>
            </div>
          </div>
        </div>

        <div class="row no-gap">
          <div class="col-md-12">
            <div class="price-carousel price-carousel-left">
              <div class="swiper">
                <div class="swiper-wrapper">
                  <div v-for="c in priceLeft" :key="c.icon" class="swiper-slide">
                    <div class="price-item">
                      <div class="icon-box">
                        <img :src="`/resources/main/images/icon-ticker-${c.icon}.svg`" alt="" />
                      </div>
                      <div class="price-body">
                        <h4>{{ c.name }}</h4>
                        <p>$12.185<span :class="c.color === 'green' ? 'price-green' : 'price-red'">21.30%</span></p>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <div class="price-carousel price-carousel-right">
              <div class="swiper">
                <div class="swiper-wrapper">
                  <div v-for="c in priceRight" :key="c.icon" class="swiper-slide">
                    <div class="price-item">
                      <div class="icon-box">
                        <img :src="`/resources/main/images/icon-ticker-${c.icon}.svg`" alt="" />
                      </div>
                      <div class="price-body">
                        <h4>{{ c.name }}</h4>
                        <p>$12.185<span :class="c.color === 'green' ? 'price-green' : 'price-red'">21.30%</span></p>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Why Choose Us -->
    <div class="why-choose-us">
      <div class="container">
        <div class="row">
          <div class="col-md-12">
            <div class="section-title">
              <h3 class="wow fadeInUp">Why Choose us ?</h3>
              <h2 class="text-anime">Know More About Signet</h2>
            </div>
          </div>
        </div>

        <div class="row">
          <div class="col-md-4">
            <div class="why-choose-us-item wow fadeInUp" data-wow-delay="0.25s">
              <div class="icon-box"><img src="/resources/main/images/icon-why-choose-us-1.svg" alt="" /></div>
              <h3>Safe & Secure</h3>
            </div>
          </div>
          <div class="col-md-4">
            <div class="why-choose-us-item wow fadeInUp" data-wow-delay="0.5s">
              <div class="icon-box"><img src="/resources/main/images/icon-why-choose-us-2.svg" alt="" /></div>
              <h3>Early Bonus</h3>
            </div>
          </div>
          <div class="col-md-4">
            <div class="why-choose-us-item wow fadeInUp" data-wow-delay="0.75s">
              <div class="icon-box"><img src="/resources/main/images/icon-why-choose-us-3.svg" alt="" /></div>
              <h3>Several Profit</h3>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Footer -->
    <footer class="main-footer">
      <div class="container">
        <div class="row">
          <div class="col-md-12">
            <div class="footer-newsletters">
              <div class="row">
                <div class="col-lg-6">
                  <div class="newsletter-title">
                    <div class="icon-box">
                      <img src="/resources/main/images/icon-stay-info.svg" alt="" />
                    </div>
                    <h2>Stay Informed And Never Miss An Signet Update!</h2>
                  </div>
                </div>
                <div class="col-lg-6">
                  <div class="newsletter-form">
                    <form id="newsletter_form" action="#" data-toggle="validator" @submit.prevent>
                      <div class="row no-gap align-items-center">
                        <div class="form-group col-md-8">
                          <input type="email" name="email" class="form-control" id="news_email" placeholder="Enter Your Email Address" required />
                          <div class="help-block with-errors"></div>
                        </div>
                        <div class="col-md-4 text-end">
                          <button type="submit" class="btn-default disabled">Subscribe Now</button>
                          <div id="news_letter_Submit" class="h3 text-left hidden"></div>
                        </div>
                      </div>
                    </form>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="row">
          <div class="col-md-12">
            <div class="mega-footer">
              <div class="row">
                <div class="col-lg-3 col-12">
                  <div class="footer-about">
                    <div class="footer-logo">
                      <img src="/resources/assets/img/brand/small-logo.png" alt="" />
                    </div>
                    <div class="footer-about-content">
                      <p>Turn $10 into success&mdash;trade smarter, earn bigger, and build your future with SIGNET!</p>
                    </div>
                    <div class="footer-social-links">
                      <ul>
                        <li><a href="#"><i class="fa-brands fa-facebook-f"></i></a></li>
                        <li><a href="#"><i class="fa-brands fa-twitter"></i></a></li>
                        <li><a href="#"><i class="fa-brands fa-instagram"></i></a></li>
                        <li><a href="#"><i class="fa-brands fa-linkedin-in"></i></a></li>
                      </ul>
                    </div>
                  </div>
                </div>

                <div class="col-lg-3 col-md-4">
                  <div class="footer-links">
                    <div class="footer-title"><h3>Quick Links</h3></div>
                    <div class="footer-menu">
                      <ul>
                        <li><a href="#">Home</a></li>
                        <li><a href="#">About Us</a></li>
                        <li><a href="#">Services</a></li>
                        <li><a href="#">Pricing</a></li>
                        <li><a href="#">Blog</a></li>
                      </ul>
                    </div>
                  </div>
                </div>

                <div class="col-lg-3 col-md-4">
                  <div class="footer-links">
                    <div class="footer-title"><h3>Extra Links</h3></div>
                    <div class="footer-menu">
                      <ul>
                        <li><a href="#">Roadmap</a></li>
                        <li><a href="#">Features</a></li>
                        <li><a href="#">Join Us</a></li>
                        <li><a href="#">Token Sale</a></li>
                        <li><a href="#">Faqs</a></li>
                      </ul>
                    </div>
                  </div>
                </div>

                <div class="col-lg-3 col-md-4">
                  <div class="footer-contact-information">
                    <div class="footer-title"><h3>Contact Information</h3></div>
                    <div class="footer-contact-info">
                      <div class="footer-contact-info-item">
                        <div class="icon-box"><i class="fa-solid fa-phone"></i></div>
                        <p>(+0) 123 456 7890</p>
                      </div>
                      <div class="footer-contact-info-item">
                        <div class="icon-box"><i class="fa-solid fa-envelope"></i></div>
                        <p>Info@Signet.com</p>
                      </div>
                      <div class="footer-contact-info-item">
                        <div class="icon-box"><i class="fa-solid fa-location-dot"></i></div>
                        <p>200 East 65th Street 17 No, Australia</p>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="row">
          <div class="col-md-12">
            <div class="footer-copyrights">
              <div class="row align-items-center">
                <div class="col-md-6">
                  <div class="footer-copyright">
                    <p>Copyright &copy; 2024 Signet. All Rights Reserved.</p>
                  </div>
                </div>
                <div class="col-md-6">
                  <div class="footer-policy-menu">
                    <ul>
                      <li><a href="#">Privacy Policy</a></li>
                      <li><a href="#">Terms of Use</a></li>
                    </ul>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </footer>
  </div>
</template>

<style scoped>
/* See the .reveal note in the top comment block: force the post-animation
   end state since the GSAP ScrollTrigger reveal script isn't loaded. */
.signet-public :deep(.reveal) {
  visibility: visible;
}
</style>
