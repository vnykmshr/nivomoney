type ClassValue = string | boolean | null | undefined | ClassValue[];

/**
 * Combines class names, filtering out falsy values.
 * Lightweight alternative to clsx/classnames.
 *
 * @example
 * cn('base', condition && 'conditional', ['array', 'of', 'classes'])
 * // => 'base conditional array of classes'
 */
export function cn(...inputs: ClassValue[]): string {
  return inputs
    .flat()
    .filter((x): x is string => typeof x === 'string' && x.length > 0)
    .join(' ');
}
