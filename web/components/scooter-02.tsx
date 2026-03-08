const Scooter02Icon = ({
  size = 24, // Increased default size for better visibility
  color = '#F8FAFC', // Cool White (Slate-50) for high contrast on dark backgrounds
  strokeWidth = 1.5,
  background = 'transparent',
  opacity = 1,
  rotation = 0,
  shadow = 0,
  flipHorizontal = false,
  flipVertical = false,
  padding = 0,
}) => {
  const transforms = []
  if (rotation !== 0) transforms.push(`rotate(${rotation}deg)`)
  // Added flipHorizontal by default to face right, or use the prop
  if (flipHorizontal) transforms.push('scaleX(-1)')
  if (flipVertical) transforms.push('scaleY(-1)')

  const viewBoxSize = 24 + padding * 2
  const viewBoxOffset = -padding
  const viewBox = `${viewBoxOffset} ${viewBoxOffset} ${viewBoxSize} ${viewBoxSize}`

  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox={viewBox}
      width={size}
      height={size}
      fill="none"
      stroke={color}
      strokeWidth={strokeWidth}
      strokeLinecap="round"
      strokeLinejoin="round"
      style={{
        opacity,
        transform: transforms.join(' ') || undefined,
        filter:
          shadow > 0
            ? `drop-shadow(0 ${shadow}px ${shadow * 2}px rgba(255,255,255,0.2))`
            : undefined,
        backgroundColor: background !== 'transparent' ? background : undefined,
      }}
    >
      <g stroke={color} strokeWidth={strokeWidth}>
        <path
          strokeLinejoin="round"
          d="M2 16c0-3.182 2.239-5 5-5s5 1.818 5 5z"
        />
        <path strokeLinecap="round" strokeLinejoin="round" d="M5 8h4" />
        <path d="M10 16a3 3 0 1 1-6 0" />
        <circle cx="20" cy="17" r="2" />
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          d="M16 8c1.333.638 4 3.174 4 7M15.99 5h.547c.984 0 1.888.58 2.344 1.503c.315.64 0 1.497-.896 1.497H15.99m0-3v3m0-3h-3.046m3.046 3c0 1.913-.212 8-3.99 8h5.666"
        />
      </g>
    </svg>
  )
}

export default Scooter02Icon
