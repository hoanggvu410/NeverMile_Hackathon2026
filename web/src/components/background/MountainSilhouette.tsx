export function MountainSilhouette() {
  return (
    <div
      aria-hidden
      className="pointer-events-none fixed inset-x-0 bottom-0 z-[1] h-[52vh]"
    >
      {/* atmospheric horizon haze behind the peaks */}
      <div
        className="absolute inset-x-0 bottom-0 h-1/2"
        style={{
          background:
            "radial-gradient(120% 100% at 50% 100%, oklch(0.45 0.12 230 / 0.28), transparent 70%)",
        }}
      />
      <svg
        viewBox="0 0 1440 520"
        preserveAspectRatio="none"
        className="h-full w-full"
      >
        <defs>
          <linearGradient id="m-far" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="oklch(0.22 0.05 280)" stopOpacity="0.55" />
            <stop offset="100%" stopColor="oklch(0.1 0.03 270)" stopOpacity="0.9" />
          </linearGradient>
          <linearGradient id="m-near" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="oklch(0.13 0.035 275)" stopOpacity="0.92" />
            <stop offset="100%" stopColor="oklch(0.06 0.02 270)" stopOpacity="1" />
          </linearGradient>
        </defs>
        <path
          fill="url(#m-far)"
          d="M0 360 L160 250 L320 330 L470 210 L640 320 L780 230 L950 330 L1120 240 L1280 320 L1440 270 L1440 520 L0 520 Z"
        />
        <path
          fill="url(#m-near)"
          d="M0 430 L180 330 L360 410 L540 300 L720 400 L900 310 L1080 410 L1260 330 L1440 400 L1440 520 L0 520 Z"
        />
      </svg>
    </div>
  );
}
