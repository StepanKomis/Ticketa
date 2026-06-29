import './HeroSection.scss'

interface Props {
  greeting: string
  subtitle: string
  className?: string
}

export default function HeroSection({ greeting, subtitle, className }: Props) {
  return (
    <section className={`heroSection${className ? ` ${className}` : ''}`}>
      <h1 className="heroSection__greeting">{greeting}</h1>
      <p className="heroSection__subtitle">{subtitle}</p>
    </section>
  )
}
