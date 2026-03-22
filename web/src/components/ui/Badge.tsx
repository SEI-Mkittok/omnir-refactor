import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from '@/lib/utils'

const badgeVariants = cva(
  'inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ring-1 ring-inset',
  {
    variants: {
      variant: {
        default: 'bg-slate-50 text-slate-700 ring-slate-600/20',
        blue: 'bg-blue-50 text-blue-700 ring-blue-600/20',
        yellow: 'bg-yellow-50 text-yellow-700 ring-yellow-600/20',
        green: 'bg-green-50 text-green-700 ring-green-600/20',
        red: 'bg-red-50 text-red-700 ring-red-600/20',
        gray: 'bg-gray-50 text-gray-600 ring-gray-500/20',
        indigo: 'bg-[#E8EDF2] text-[#1B3A4B] ring-[#1B3A4B]/20',
        purple: 'bg-teal-50 text-teal-700 ring-teal-600/20',
        orange: 'bg-orange-50 text-orange-700 ring-orange-600/20',
      },
    },
    defaultVariants: {
      variant: 'default',
    },
  }
)

export interface BadgeProps
  extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof badgeVariants> {}

export function Badge({ className, variant, ...props }: BadgeProps) {
  return <div className={cn(badgeVariants({ variant }), className)} {...props} />
}

export { badgeVariants }
