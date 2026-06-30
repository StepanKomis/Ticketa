import { useState, useEffect, useRef } from 'react'
import './KebabMenu.css'

type KebabAction = {
  type: 'action'
  label: string
  onClick: () => void
  disabled?: boolean
  danger?: boolean
}

type KebabSubmenu = {
  type: 'submenu'
  label: string
  items: Array<{
    label: string
    onClick: () => void
    checked?: boolean
    disabled?: boolean
  }>
}

export type KebabItem = KebabAction | KebabSubmenu

interface Props {
  ariaLabel: string
  items: KebabItem[]
}

export default function KebabMenu({ ariaLabel, items }: Props) {
  const [open, setOpen] = useState(false)
  const [above, setAbove] = useState(false)
  const [openSubmenu, setOpenSubmenu] = useState<number | null>(null)
  const ref = useRef<HTMLDivElement>(null)
  const subTimer = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => {
    if (!open) { setOpenSubmenu(null); return }
    const onOutside = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false)
    }
    document.addEventListener('mousedown', onOutside)
    return () => document.removeEventListener('mousedown', onOutside)
  }, [open])

  // Cleanup timer on unmount
  useEffect(() => () => { if (subTimer.current) clearTimeout(subTimer.current) }, [])

  const clearSubTimer = () => { if (subTimer.current) { clearTimeout(subTimer.current); subTimer.current = null } }

  const toggle = (e: React.MouseEvent<HTMLButtonElement>) => {
    if (!open) {
      const rect = e.currentTarget.getBoundingClientRect()
      setAbove(window.innerHeight - rect.bottom < 200)
    }
    setOpen(o => !o)
  }

  const close = () => { setOpen(false); setOpenSubmenu(null) }

  const handleSubEnter = (i: number) => { clearSubTimer(); setOpenSubmenu(i) }
  const handleSubLeave = () => { subTimer.current = setTimeout(() => setOpenSubmenu(null), 150) }
  const handleOtherEnter = () => { clearSubTimer(); setOpenSubmenu(null) }

  return (
    <div className="kebabMenu" ref={ref}>
      <button
        type="button"
        className="kebabMenu__trigger"
        aria-label={ariaLabel}
        aria-haspopup="menu"
        aria-expanded={open}
        onClick={toggle}
      >
        <svg width="16" height="16" viewBox="0 0 16 16" fill="none" aria-hidden="true">
          <circle cx="8" cy="3.5" r="1.2" fill="currentColor" />
          <circle cx="8" cy="8" r="1.2" fill="currentColor" />
          <circle cx="8" cy="12.5" r="1.2" fill="currentColor" />
        </svg>
      </button>

      {open && (
        <ul
          className={`kebabMenu__list${above ? ' kebabMenu__list--above' : ''}`}
          role="menu"
        >
          {items.map((item, i) => {
            if (item.type === 'action') {
              return (
                <li role="none" key={i} onMouseEnter={handleOtherEnter}>
                  <button
                    type="button"
                    role="menuitem"
                    className={`kebabMenu__item${item.danger ? ' kebabMenu__item--danger' : ''}`}
                    disabled={item.disabled}
                    onClick={() => { item.onClick(); close() }}
                  >
                    {item.label}
                  </button>
                </li>
              )
            }
            const isOpen = openSubmenu === i
            return (
              <li
                role="none"
                key={i}
                className="kebabMenu__itemSub"
                onMouseEnter={() => handleSubEnter(i)}
                onMouseLeave={handleSubLeave}
              >
                <button
                  type="button"
                  role="menuitem"
                  aria-haspopup="menu"
                  aria-expanded={isOpen}
                  className={`kebabMenu__item kebabMenu__item--hasSub${isOpen ? ' kebabMenu__item--subOpen' : ''}`}
                  onClick={() => { clearSubTimer(); setOpenSubmenu(isOpen ? null : i) }}
                >
                  {item.label}
                  <svg
                    width="12" height="12" viewBox="0 0 12 12" fill="none"
                    aria-hidden="true"
                    className={`kebabMenu__arrow${isOpen ? ' kebabMenu__arrow--open' : ''}`}
                  >
                    <path d="M4 2.5L7.5 6 4 9.5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
                  </svg>
                </button>
                {isOpen && (
                  <ul
                    className="kebabMenu__submenu"
                    role="menu"
                    onMouseEnter={clearSubTimer}
                    onMouseLeave={handleSubLeave}
                  >
                    {item.items.map((sub, j) => (
                      <li role="none" key={j}>
                        <button
                          type="button"
                          role="menuitemradio"
                          aria-checked={sub.checked}
                          className={`kebabMenu__item${sub.checked ? ' kebabMenu__item--current' : ''}`}
                          disabled={sub.disabled ?? sub.checked}
                          onClick={() => { sub.onClick(); close() }}
                        >
                          {sub.checked && <span className="kebabMenu__check" aria-hidden="true">✓</span>}
                          {sub.label}
                        </button>
                      </li>
                    ))}
                  </ul>
                )}
              </li>
            )
          })}
        </ul>
      )}
    </div>
  )
}
