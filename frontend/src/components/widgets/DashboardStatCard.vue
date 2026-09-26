<script setup>
// One stat tile from the Admin/Agent/User dashboard.blade.php grids. The
// Blade views repeat this markup per stat with small variations, all
// exposed as props:
//  - `primary`: the first "My Token" tile — icon-shape-primary with the
//    people icon, and a plain `h2.h5` / `h3.fw-extrabold.mb-1` mobile
//    heading (every other tile uses `h2.fw-extrabold.h5` / `h3.mb-1`).
//  - `usdt`: appends the green " USDT" suffix to both readouts.
// The default slot renders under the desktop readout (used by the
// "This Time My Share Portion" tile's "This Time My Share Value" line).
defineProps({
  colClass: { type: String, default: 'col-12 col-sm-12 col-xl-4 mb-4' },
  label: { type: String, required: true },
  value: { type: [String, Number], default: '' },
  usdt: { type: Boolean, default: false },
  primary: { type: Boolean, default: false },
})
</script>

<template>
  <div :class="colClass">
    <div class="card border-0 shadow">
      <div class="card-body">
        <div class="row d-block d-xl-flex align-items-center">
          <div class="col-12 col-xl-5 text-xl-center mb-3 mb-xl-0 d-flex align-items-center justify-content-xl-center">
            <div v-if="primary" class="icon-shape icon-shape-primary rounded me-4 me-sm-0">
              <svg class="icon" fill="currentColor" viewBox="0 0 20 20" xmlns="http://www.w3.org/2000/svg"><path d="M13 6a3 3 0 11-6 0 3 3 0 016 0zM18 8a2 2 0 11-4 0 2 2 0 014 0zM14 15a4 4 0 00-8 0v3h8v-3zM6 8a2 2 0 11-4 0 2 2 0 014 0zM16 18v-3a5.972 5.972 0 00-.75-2.906A3.005 3.005 0 0119 15v3h-3zM4.75 12.094A5.973 5.973 0 004 15v3H1v-3a3 3 0 013.75-2.906z"></path></svg>
            </div>
            <div v-else class="icon-shape icon-shape-secondary rounded me-4 me-sm-0">
              <svg class="icon" fill="currentColor" viewBox="0 0 20 20" xmlns="http://www.w3.org/2000/svg"><path fill-rule="evenodd" d="M10 2a4 4 0 00-4 4v1H5a1 1 0 00-.994.89l-1 9A1 1 0 004 18h12a1 1 0 00.994-1.11l-1-9A1 1 0 0015 7h-1V6a4 4 0 00-4-4zm2 5V6a2 2 0 10-4 0v1h4zm-6 3a1 1 0 112 0 1 1 0 01-2 0zm7-1a1 1 0 100 2 1 1 0 000-2z" clip-rule="evenodd"></path></svg>
            </div>
            <div class="d-sm-none">
              <h2 :class="primary ? 'h5' : 'fw-extrabold h5'">{{ label }}</h2>
              <h3 :class="primary ? 'fw-extrabold mb-1' : 'mb-1'">{{ value }}<span v-if="usdt" class="text-success fw-bold fs-5"> USDT</span></h3>
            </div>
          </div>
          <div class="col-12 col-xl-7 px-xl-0">
            <div class="d-none d-sm-block">
              <h2 class="h6 text-gray-400 mb-0">{{ label }}</h2>
              <h3 class="fw-extrabold mb-2">{{ value }}<span v-if="usdt" class="text-success fw-bold fs-5"> USDT</span></h3>
            </div>
            <slot />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
