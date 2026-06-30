import type { ReactNode } from 'react'
import './Card.css'

interface CardProps {
  title?: string
  className?: string
  children: ReactNode
}

export default function Card({ title, className, children }: CardProps) {
  return (
    <div className={`card${className ? ` ${className}` : ''}`}>
      {title && <h2 className="card__title">{title}</h2>}
      {children}
    </div>
  )
}
