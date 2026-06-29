import './Card.scss'

type CardProps = {
  title?: string
  children: React.ReactNode
} & React.HTMLAttributes<HTMLElement>

export default function Card({ title, className, children, ...props }: CardProps) {
  return (
    <section className={['card', className].filter(Boolean).join(' ')} {...props}>
      {title && <h2 className="card__title">{title}</h2>}
      {children}
    </section>
  )
}
