package place

import (
	"fmt"
	"math"
	"os"
)

/*===================================================================*/
//  
//     place_qpsolver.c
//
//        Philip Chong
//              pchong@cadence.com
//
/*===================================================================*/

//#include <assert.h>
//#include <math.h>
//#include <stdio.h>
//#include <stdlib.h>

//#include "place_qpsolver.h"

//ABC_NAMESPACE_IMPL_START


#undef  QPS_DEBUG

const QPS_TOL = 1.0e-3
#define QPS_EPS (QPS_TOL * QPS_TOL)

const QPS_MAX_TOL = 0.1
const QPS_LOOP_TOL = 1.0e-3

const QPS_RELAX_ITER = 180
const QPS_MAX_ITER = 200
const QPS_STEPSIZE_RETRIES = 2
const QPS_MINSTEP = 1.0e-6
const QPS_DEC_CHANGE = 0.01

#define QPS_PRECON
const QPS_PRECON_EPS = 1.0e-9

#undef QPS_HOIST

                      
                                           
      

     
                                                                         
                                                                      
                                                                          
                                                                         
                                                                              
                                            
        

                                                                        
                                                                            
                                           
        

                                                                             
                                        
      

                                                                  
                                                           

      

/**********************************************************************/

func qps_settp(p *qps_problem_t) {
  /* Fill in the p->priv_tp array with the current locations of all cells
     (independent, dependent and fixed). */

  var i int32;
  var t, u int32;
  var pr int32;
  var rx, ry qps_float_t;
  var ta qps_float_t;

  var ii *int32  = p.priv_ii;
  var tp *qps_float_t  = p.priv_tp;
  var cp *qps_float_t  = p.priv_cp;

  /* do independent and fixed cells first */
  for i = p.num_cells; i--; {
    t = ii[i];
    if t >= 0 {        /* indep cell */
      tp[i * 2] = cp[t];
      tp[i * 2 + 1] = cp[t + 1];
    
    } else if t == -1 {        /* fixed cell */
      tp[i * 2] = p.x[i];
      tp[i * 2 + 1] = p.y[i];
    }
  }
  /* now do dependent cells */
  for i = p.num_cells; i--; {
    if ii[i] < -1 {
      t = -(ii[i] + 2);        /* index of COG constraint */
      ta = 0.0;
      rx = 0.0;
      ry = 0.0;
      pr = p.priv_gt[t];
      for (u = p.cog_list[pr++]) >= 0 {
    ta += p.area[u];
    if u != i {
      rx -= p.area[u] * tp[u * 2];
      ry -= p.area[u] * tp[u * 2 + 1];
    }
      }
      rx += p.cog_x[t] * ta;
      ry += p.cog_y[t] * ta;
      tp[i * 2] = rx / p.area[i];
      tp[i * 2 + 1] = ry / p.area[i];
    }
  }

                   
                                           
                                      
                                                             
   
      
}

/**********************************************************************/

func qps_func(p *qps_problem_t) qps_float_t {
  /* Return f(p).  qps_settp() should have already been called before
     entering here */

  var j, k int32;
  var pr int32;
  var jx, jy, tx, ty qps_float_t;
  var f qps_float_t;
  var w qps_float_t;

                       
  var i int32;
  var st int32;
  var kx, ky, sx, sy qps_float_t;
  var t qps_float_t;
      

  var tp *qps_float_t  = p.priv_tp;

  f = 0.0;
  pr = 0;
  for j = 0; j < p.num_cells; j++ {
    jx = tp[j * 2];
    jy = tp[j * 2 + 1];
    for (k = p.priv_cc[pr]) >= 0 {
      w = p.priv_cw[pr];
      tx = tp[k * 2] - jx;
      ty = tp[k * 2 + 1] - jy;
      f += w * (tx * tx + ty * ty);
      pr++;
    }
    pr++;
  }
  p.f = f;

                       
  /* loop penalties */
  pr = 0;
  for i = 0; i < p.loop_num; i++ {
    t = 0.0;
    j = st = p.loop_list[pr++];
    jx = sx = tp[j * 2];
    jy = sy = tp[j * 2 + 1];
    for (k = p.loop_list[pr]) >= 0 {
      kx = tp[k * 2];
      ky = tp[k * 2 + 1];
      tx = jx - kx;
      ty = jy - ky;
      t += tx * tx + ty * ty;
      j = k;
      jx = kx;
      jy = ky;
      pr++;
    }
    tx = jx - sx;
    ty = jy - sy;
    t += tx * tx + ty * ty;
    t -= p.loop_max[i];
                   
                                                       
                              
      
    p.priv_lt[i] = t;
    f += p.loop_penalty[i] * t;
    pr++;
  }
                      

  if p.max_enable {
    for j = p.num_cells; j--; {
      f += p.priv_mxl[j] * (-tp[j * 2]);
      f += p.priv_mxh[j] * (tp[j * 2] - p.max_x);
      f += p.priv_myl[j] * (-tp[j * 2 + 1]);
      f += p.priv_myh[j] * (tp[j * 2 + 1] - p.max_y);
    }
  }

                   
                                                         
      
  return f;
}

/**********************************************************************/

func qps_dfunc(p *qps_problem_t, d *qps_float_t) {
  /* Set d to grad f(p).  First computes partial derivatives wrt all cells
     then finds gradient wrt only the independent cells.  qps_settp() should
     have already been called before entering here */

  var i, j, k int32;
  var pr int32;
  var jx, jy, kx, ky, tx, ty qps_float_t;
  var ji, ki int32;
  var w qps_float_t;

                       
  var sx, sy qps_float_t;
  var st int32;
      

  var tp *qps_float_t  = p.priv_tp;
  var tp2 *qps_float_t  = p.priv_tp2;

  /* compute partials and store in tp2 */
  for i = p.num_cells; i--; {
    tp2[i * 2] = 0.0;
    tp2[i * 2 + 1] = 0.0;
  }
  for j = 0; j < p.num_cells; j++ {
    jx = tp[j * 2];
    jy = tp[j * 2 + 1];
    for (k = p.priv_cc[pr]) >= 0 {
      w = 2.0 * p.priv_cw[pr];
      kx = tp[k * 2];
      ky = tp[k * 2 + 1];
      tx = w * (jx - kx);
      ty = w * (jy - ky);
      tp2[j * 2] += tx;
      tp2[k * 2] -= tx;
      tp2[j * 2 + 1] += ty;
      tp2[k * 2 + 1] -= ty;
      pr++;
    }
    pr++;
  }

                       
  /* loop penalties */
  pr = 0;
  for i = 0; i < p.loop_num; i++ {
    j = st = p.loop_list[pr++];
    jx = sx = tp[j * 2];
    jy = sy = tp[j * 2 + 1];
    w = 2.0 * p.loop_penalty[i];
    for (k = p.loop_list[pr]) >= 0 {
      kx = tp[k * 2];
      ky = tp[k * 2 + 1];
      tx = w * (jx - kx);
      ty = w * (jy - ky);
      tp2[j * 2] += tx;
      tp2[k * 2] -= tx;
      tp2[j * 2 + 1] += ty;
      tp2[k * 2 + 1] -= ty;
      j = k;
      jx = kx;
      jy = ky;
      pr++;
    }
    tx = w * (jx - sx);
    ty = w * (jy - sy);
    tp2[j * 2] += tx;
    tp2[st * 2] -= tx;
    tp2[j * 2 + 1] += ty;
    tp2[st * 2 + 1] -= ty;
    pr++;
  }
                      

  if p.max_enable {
    for j = p.num_cells; j--; {
      tp2[j * 2] += p.priv_mxh[j] - p.priv_mxl[j];
      tp2[j * 2 + 1] += p.priv_myh[j] - p.priv_myl[j];
    }
  }

                   
                                                    
                                      
                                                               
   
      

  /* translate partials to independent variables */
  for j = p.priv_n; j--; {
    d[j] = 0.0;
  }
  for j = p.num_cells; j--; {
    ji = p.priv_ii[j];
    if ji >= 0 {        /* indep var */
      d[ji] += tp2[j * 2];
      d[ji + 1] += tp2[j * 2 + 1];
    
    } else if ji < -1 {        /* dependent variable */
      ji = -(ji + 2);        /* get COG index */
      pr = p.priv_gt[ji];
      for (k = p.cog_list[pr]) >= 0 {
    ki = p.priv_ii[k];
    if ki >= 0 {
      w = p.priv_gw[pr];
                   
                                                         
      
      d[ki] -= tp2[j * 2] * w;
      d[ki + 1] -= tp2[j * 2 + 1] * w;
    }
    pr++;
      }
    }
  }

                   
                                                    
                                   
                                      
   
      
}

/**********************************************************************/

func qps_linmin(p *qps_problem_t, dgg qps_float_t, h *qps_float_t) {
  /* Perform line minimization.  p->priv_cp is the current location, h is
     direction of the gradient.  Updates p->priv_cp to line minimal position
     based on formulas from "Handbook of Applied Optimization", Pardalos and
     Resende, eds., Oxford Univ. Press, 2002.  qps_settp() should have
     already been called before entering here.  Since p->priv_cp is changed,
     p->priv_tp array becomes invalid following this routine. */

  var i, j, k int32;
  var pr int32;
  var ji, ki int32;
  var jx, jy, kx, ky qps_float_t;
  var f qps_float_t  = 0.0;
  var w qps_float_t;

                       
  var st int32;
  var sx, sy, tx, ty qps_float_t;
  var t qps_float_t;
      

  var tp *qps_float_t  = p.priv_tp;

  /* translate h vector to partials over all variables and store in tp */
  for i = p.num_cells; i--; {
    tp[i * 2] = 0.0;
    tp[i * 2 + 1] = 0.0;
  }
  for j = p.num_cells; j--; {
    ji = p.priv_ii[j];
    if ji >= 0 {        /* indep cell */
      tp[j * 2] = h[ji];
      tp[j * 2 + 1] = h[ji + 1];
    
    } else if ji < -1 {        /* dep cell */
      ji = -(ji + 2);        /* get COG index */
      pr = p.priv_gt[ji];
      for (k = p.cog_list[pr]) >= 0 {
    ki = p.priv_ii[k];
    if ki >= 0 {
      w = p.priv_gw[pr];
                   
                                                         
      
      tp[j * 2] -= h[ki] * w;
      tp[j * 2 + 1] -= h[ki + 1] * w;
    }
    pr++;
      }
    }
  }

  /* take product x^T Z^T C Z x */
  pr = 0;
  for j = 0; j < p.num_cells; j++ {
    jx = tp[j * 2];
    jy = tp[j * 2 + 1];
    for (k = p.priv_cc[pr]) >= 0 {
      w = p.priv_cw[pr];
      kx = tp[k * 2] - jx;
      ky = tp[k * 2 + 1] - jy;
      f += w * (kx * kx + ky * ky);
      pr++;
    }
    pr++;
  }

                       
  /* add loop penalties */
  pr = 0;
  for i = 0; i < p.loop_num; i++ {
    t = 0.0;
    j = st = p.loop_list[pr++];
    jx = sx = tp[j * 2];
    jy = sy = tp[j * 2 + 1];
    for (k = p.loop_list[pr]) >= 0 {
      kx = tp[k * 2];
      ky = tp[k * 2 + 1];
      tx = jx - kx;
      ty = jy - ky;
      t += tx * tx + ty * ty;
      j = k;
      jx = kx;
      jy = ky;
      pr++;
    }
    tx = jx - sx;
    ty = jy - sy;
    t += tx * tx + ty * ty;
    f += p.loop_penalty[i] * t;
    pr++;
  }
                      

                   
            
      

  /* compute step size */
  f = (dgg / f) / 2.0;
  for j = p.priv_n; j--; {
    p.priv_cp[j] += f * h[j];
  }
                   
                                                       
                                   
                                               
   
      
}

/**********************************************************************/

func qps_cgmin(p *qps_problem_t) {
  /* Perform CG minimization. Mostly from "Numerical Recipes", Press et al.,
     Cambridge Univ. Press, 1992, with some changes to help performance in
     our restricted problem domain. */

  var fp, gg, dgg, gam qps_float_t;
  var t qps_float_t;
  var i, j int32;

  var n int32  = p.priv_n;
  var g *qps_float_t  = p.priv_g;
  var h *qps_float_t  = p.priv_h;
  var xi *qps_float_t  = p.priv_xi;

  qps_settp(p);
  fp = qps_func(p);
  qps_dfunc(p, g);

  dgg = 0.0;
  for j = n; j--; {
    g[j] = -g[j];
    h[j] = g[j];
                       
    h[j] *= p.priv_pcgt[j];
      
    dgg += g[j] * h[j];
  }

  for i = 0; i < 2 * n; i++ {

                   
                                                 
                                     
                                                 
     
      

    if dgg == 0.0 {
      break;
    }
    qps_linmin(p, dgg, h);
    qps_settp(p);
    p.priv_f = qps_func(p);
    if math.Abs(float64((p.priv_f) - fp)) <=
    (math.Abs(float64(p.priv_f)) + math.Abs(float64(fp)) + QPS_EPS) * QPS_TOL / 2.0 {
      break;
    }
    fp = p.priv_f;
    qps_dfunc(p, xi);
    gg = dgg;
    dgg = 0.0;
    for j = n; j--; {
      t = xi[j] * xi[j];
                       
      t *= p.priv_pcgt[j];
      
      dgg += t;
    }
    gam = dgg / gg;
    for j = n; j--; {
      g[j] = -xi[j];
      t = g[j];
                       
      t *= p.priv_pcgt[j];
      
      h[j] = t + gam * h[j];
    }
  }
                   
                                                                             
      
  if i == 2 * n {
    fmt.Fprintf(os.Stderr, "### Too many iterations in qps_cgmin()\n");
                      
                                                                    
      
  }
}

/**********************************************************************/

func qps_init(p *qps_problem_t) {
  var i, j int32;
  var pr, pw int32;

                      
                                          
                     
      

                   
                                                                         
                   
         
                                            
                                      
                                       
                                 
                                                                          
           
     
                                       
         
   
                                              
                                      
                                                                     
                     
   
     
                   
                                           
                                        
                                                            
     
   
         
                                         
                                    
                                       
                                  
                                                  
           
     
                                
         
   
                                              
                                    
                                                                          
   
      
                      

  p.priv_ii = (int *)malloc(p.num_cells * sizeof(int));
  assert(p.priv_ii, "p.priv_ii");

  p.max_enable = 0;

  p.priv_fopt = 0.0;

  /* canonify c and w */
  pr = pw = 0;
  for i = 0; i < p.num_cells; i++ {
    for (j = p.connect[pr]) >= 0 {
      if j > i {
    pw++;
      }
      pr++;
    }
    pw++;
    pr++;
  }
  p.priv_cc = (int *)malloc(pw * sizeof(int));
  assert(p.priv_cc, "p.priv_cc");
  p.priv_cr = (int *)malloc(p.num_cells * sizeof(int));
  assert(p.priv_cr, "p.priv_cr");
  p.priv_cw = (qps_float_t*)malloc(pw * sizeof(qps_float_t));
  assert(p.priv_cw, "p.priv_cw");
  p.priv_ct = (qps_float_t*)malloc(pw * sizeof(qps_float_t));
  assert(p.priv_ct, "p.priv_ct");
  p.priv_cm = pw;
  pr = pw = 0;
  for i = 0; i < p.num_cells; i++ {
    p.priv_cr[i] = pw;
    for (j = p.connect[pr]) >= 0 {
      if j > i {
    p.priv_cc[pw] = p.connect[pr];
    p.priv_ct[pw] = p.edge_weight[pr];
    pw++;
      }
      pr++;
    }
    p.priv_cc[pw] = -1;
    p.priv_ct[pw] = -1.0;
    pw++;
    pr++;
  }
  assert(pw == p.priv_cm, "pw == p.priv_cm");

  /* temp arrays for function eval */
  p.priv_tp = (qps_float_t *) malloc(4 * p.num_cells * sizeof(qps_float_t));
  assert(p.priv_tp, "p.priv_tp");
  p.priv_tp2 = p.priv_tp + 2 * p.num_cells;
}

/**********************************************************************/

func qps_estopt(p *qps_problem_t) qps_float_t {
  var i, j, cell int32;
  var r qps_float_t;
  var t1, t2 *qps_float_t;
  var t qps_float_t;

  if p.max_enable {
    r = 0.0;
    t1 = (qps_float_t *) malloc(2 * p.num_cells * sizeof(qps_float_t));
                   
               
      
    for i = 2 * p.num_cells; i--; {
      t1[i] = 0.0;
    }
    j = 0;
    for i = 0; i < p.cog_num; i++ {
      for (cell = p.cog_list[j]) >= 0 {
    t1[cell * 2] = p.cog_x[i];
    t1[cell * 2 + 1] = p.cog_y[i];
    j++;
      }
      j++;
    }
    t2 = p.priv_tp;
    p.priv_tp = t1;
    r = qps_func(p);
    p.priv_tp = t2;
    free(t1);
    t = (p.max_x * p.max_x + p.max_y * p.max_y);
    t *= p.num_cells;
    for i = p.num_cells; i--; {
      if p.fixed[i] {
    r += t;
      }
    }
  
  } else {
    r = p.priv_f;
  }
  if p.loop_num {
    /* FIXME hacky */
    r *= 8.0;
  }
  return r;
}

/**********************************************************************/

func qps_solve_inner(p *qps_problem_t) {
  var i int32;
  var t qps_float_t;
  var z qps_float_t;
  var pm1, pm2, tp qps_float_t;
  var tw *qps_float_t;
                      
           
                                             
             
      

  tw = p.priv_cw;
                      
                     
                            
   
        
                             
                                    
     
                                     
           
                                       
                                            
                 
                                          
     
       
           
     
   
                      
  p.priv_cw = p.priv_ct;
                      

  qps_cgmin(p);

  if p.max_enable || p.loop_num {
    if p.max_enable == 1 || (p.loop_num && p.loop_k == 0) {
      p.priv_eps = 2.0;
      p.priv_fmax = p.priv_f;
      p.priv_fprev = p.priv_f;
      p.priv_fopt = qps_estopt(p);
      p.priv_pn = 0;
      p.loop_fail = 0;
    
    } else {
      if p.priv_f < p.priv_fprev &&
      (p.priv_fprev - p.priv_f) >
      QPS_DEC_CHANGE * fabs(p.priv_fprev) {
    if p.priv_pn++ >= QPS_STEPSIZE_RETRIES {
      p.priv_eps /= 2.0;
      p.priv_pn = 0;
    }
      }
      p.priv_fprev = p.priv_f;
      if p.priv_fmax < p.priv_f {
    p.priv_fmax = p.priv_f;
      }
      if p.priv_f >= p.priv_fopt {
    p.priv_fopt = p.priv_fmax * 2.0;
    p.loop_fail |= 2;
                   
                                                        
      
      }
    }
                   
                                                             
                                                            
                       
      
  }

  p.loop_done = 1;
  if p.loop_num {
                   
                                                            
      
    p.loop_k++;

                      
                             
           
                                       
              
                                  
                                  
                                      
                                           
                           
                               
                 
                 
                           
          
            
            
         
       
                   
                   
                             
                                         
           
     
                      

    /* check KKT conditions */
                   
                                 
                                      
                                                                        
       
     
      
    t = 0.0;
    for i = p.loop_num; i--; {
      if p.priv_lt[i] > 0.0 || p.loop_penalty[i] > 0.0 {
    t += p.priv_lt[i] * p.priv_lt[i];
      }
      if math.Abs(float64(p.priv_lt[i])) < QPS_LOOP_TOL {
                   
                                                              
      
    continue;
      }
      z = QPS_LOOP_TOL * p.loop_max[i];
      if p.priv_lt[i] > z || (p.loop_k < QPS_RELAX_ITER &&
                p.loop_penalty[i] * p.priv_lt[i] < -z) {
    p.loop_done = 0;
                   
                                                           
                                                              
      
      }
                   
            
                                                       
                                                              
       
      
    }
    /* update penalties */
    if !p.loop_done {
      t = p.priv_eps * (p.priv_fopt - p.priv_f) / t;
      tp = 0.0;
      for i = p.loop_num; i--; {
    pm1 = p.loop_penalty[i];
                   
                                                                      
                                                                  
      
    p.loop_penalty[i] += t * p.priv_lt[i];
    if p.loop_penalty[i] < 0.0 {
      p.loop_penalty[i] = 0.0;
    }
    pm2 = p.loop_penalty[i];
    tp += math.Abs(float64(pm1 - pm2));
      }
                   
                                                      
      
    }
  }

  p.max_done = 1;
  if p.max_enable {
                   
                                                                    
      
    t = 0.0;
    for i = p.num_cells; i--; {
      z = -(p.x[i]);
      t += z * z;
      if z > QPS_TOL || (p.max_enable < QPS_RELAX_ITER &&
              p.priv_mxl[i] * z < -QPS_MAX_TOL) {
    p.max_done = 0;
                   
                                                                    
      
      }
      z = (p.x[i] - p.max_x);
      t += z * z;
      if z > QPS_TOL || (p.max_enable < QPS_RELAX_ITER &&
              p.priv_mxh[i] * z < -QPS_MAX_TOL) {
    p.max_done = 0;
                   
                                                                    
      
      }
      z = -(p.y[i]);
      t += z * z;
      if z > QPS_TOL || (p.max_enable < QPS_RELAX_ITER &&
              p.priv_myl[i] * z < -QPS_MAX_TOL) {
    p.max_done = 0;
                   
                                                                    
      
      }
      z = (p.y[i] - p.max_y);
      t += z * z;
      if z > QPS_TOL || (p.max_enable < QPS_RELAX_ITER &&
              p.priv_myh[i] * z < -QPS_MAX_TOL) {
    p.max_done = 0;
                   
                                                                    
      
      }
    }
                   
                                                                
      
    if !p.max_done {
      t = p.priv_eps * (p.priv_fopt - p.priv_f) / t;
      tp = 0.0;
      for i = p.num_cells; i--; {
    z = -(p.x[i]);
    pm1 = p.priv_mxl[i];
    p.priv_mxl[i] += t * z;
    if p.priv_mxl[i] < 0.0 {
      p.priv_mxl[i] = 0.0;
    }
    pm2 = p.priv_mxl[i];
    tp += math.Abs(float64(pm1 - pm2));

    z = (p.x[i] - p.max_x);
    pm1 = p.priv_mxh[i];
    p.priv_mxh[i] += t * z;
    if p.priv_mxh[i] < 0.0 {
      p.priv_mxh[i] = 0.0;
    }
    pm2 = p.priv_mxh[i];
    tp += math.Abs(float64(pm1 - pm2));

    z = -(p.y[i]);
    pm1 = p.priv_myl[i];
    p.priv_myl[i] += t * z;
    if p.priv_myl[i] < 0.0 {
      p.priv_myl[i] = 0.0;
    }
    pm2 = p.priv_myl[i];
    tp += math.Abs(float64(pm1 - pm2));

    z = (p.y[i] - p.max_y);
    pm1 = p.priv_myh[i];
    p.priv_myh[i] += t * z;
    if p.priv_myh[i] < 0.0 {
      p.priv_myh[i] = 0.0;
    }
    pm2 = p.priv_myh[i];
    tp += math.Abs(float64(pm1 - pm2));
      }
    }
                   
                                  
                                                                
                                                                          
     
      
    p.max_enable++;
  }

  if p.loop_k >= QPS_MAX_ITER || p.priv_eps < QPS_MINSTEP {
    p.loop_fail |= 1;
  }

  if p.loop_fail {
    p.loop_done = 1;
  }

  p.priv_cw = tw;
}

/**********************************************************************/

func qps_solve(p *qps_problem_t) {
  var i, j int32;
  var pr, pw int32;
  var bk qps_float_t;
  var tk int32;

                       
  var c int32;
  var t qps_float_t;
      

                      
        
         
             
      

  if p.max_enable {
    p.priv_mxl = (qps_float_t *)
    malloc(4 * p.num_cells * sizeof(qps_float_t));
    assert(p.priv_mxl, "p.priv_mxl");
    p.priv_mxh = p.priv_mxl + p.num_cells;
    p.priv_myl = p.priv_mxl + 2 * p.num_cells;
    p.priv_myh = p.priv_mxl + 3 * p.num_cells;
    for i = 4 * p.num_cells; i--; {
      p.priv_mxl[i] = 0.0;
    }
  }

  /* flag fixed cells with -1 */
  for i = p.num_cells; i--; {
    p.priv_ii[i] = (p.fixed[i]) ? (-1) : (0);
  }

  /* read gl and set up dependent variables */
  if p.cog_num {
    p.priv_gt = (int *)malloc(p.cog_num * sizeof(int));
    assert(p.priv_gt, "p.priv_gt");
    p.priv_gm = (qps_float_t*)malloc(p.cog_num * sizeof(qps_float_t));
    assert(p.priv_gm, "p.priv_gm");
    pr = 0;
    for i = 0; i < p.cog_num; i++ {
      tk = -1;
      bk = -1.0;
      pw = pr;
      for (j = p.cog_list[pr++]) >= 0 {
    if !p.fixed[j] {
      /* use largest entry for numerical stability; see Gordian paper */
      if p.area[j] > bk {
        tk = j;
        bk = p.area[j];
      }
    }
      }
      assert(bk > 0.0, "bk > 0.0");
      /* dependent variables have index=(-2-COG_constraint) */
      p.priv_ii[tk] = -2 - i;
      p.priv_gt[i] = pw;
      p.priv_gm[i] = bk;
    }
    p.priv_gw = (qps_float_t*)malloc(pr * sizeof(qps_float_t));
    assert(p.priv_gw, "p.priv_gw");
    pr = 0;
    for i = 0; i < p.cog_num; i++ {
      for (j = p.cog_list[pr]) >= 0 {
    p.priv_gw[pr] = p.area[j] / p.priv_gm[i];
    pr++;
      }
      p.priv_gw[pr] = -1.0;
      pr++;
    }
  }

  /* set up indexes from independent floating cells to variables */
  p.priv_n = 0;
  for i = p.num_cells; i--; {
    if !p.priv_ii[i] {
      p.priv_ii[i] = 2 * (p.priv_n++);
    }
  }
  p.priv_n *= 2;

                   
                                      
                                                            
   
      

                       
  p.priv_pcg = (qps_float_t *) malloc(p.num_cells * sizeof(qps_float_t));
  assert(p.priv_pcg, "p.priv_pcg");
  p.priv_pcgt = (qps_float_t *) malloc(p.priv_n * sizeof(qps_float_t));
  assert(p.priv_pcgt, "p.priv_pcgt");
  for i = p.num_cells; i--; {
    p.priv_pcg[i] = 0.0;
  }
  pr = 0;
  for i = 0; i < p.num_cells; i++ {
    for (c = p.priv_cc[pr]) >= 0 {
      t = p.priv_ct[pr];
      p.priv_pcg[i] += t;
      p.priv_pcg[c] += t;
      pr++;
    }
    pr++;
  }
  pr = 0;
  for i = 0; i < p.loop_num; i++ {
    t = 2.0 * p.loop_penalty[i];
    for (c = p.loop_list[pr++]) >= 0 {
      p.priv_pcg[c] += t;
    }
    pr++;
  }
                   
                                
                                                                   
   
      
  for i = p.priv_n; i--; {
      p.priv_pcgt[i] = 0.0;
  }
  for i = 0; i < p.num_cells; i++ {
    c = p.priv_ii[i];
    if c >= 0 {
      t = p.priv_pcg[i];
      p.priv_pcgt[c] += t;
      p.priv_pcgt[c + 1] += t;
    }
     
                      
                              
                                            
                       
                  
                                  
                                 
                            
                                
     
       
     
      
  }
  for i = 0; i < p.priv_n; i++ {
    t = p.priv_pcgt[i];
    if math.Abs(float64(t)) < QPS_PRECON_EPS || math.Abs(float64(t)) > 1.0/QPS_PRECON_EPS {
      p.priv_pcgt[i] = 1.0;
    
    } else {
      p.priv_pcgt[i] = 1.0 / p.priv_pcgt[i];
    }
  }
      

  /* allocate variable storage */
  p.priv_cp = (qps_float_t *) malloc(4 * p.priv_n * sizeof(qps_float_t));
  assert(p.priv_cp, "p.priv_cp");

  /* temp arrays for cg */
  p.priv_g = p.priv_cp + p.priv_n;
  p.priv_h = p.priv_cp + 2 * p.priv_n;
  p.priv_xi = p.priv_cp + 3 * p.priv_n;

  /* set values */
  for i = p.num_cells; i--; {
    if p.priv_ii[i] >= 0 {
      p.priv_cp[p.priv_ii[i]] = p.x[i];
      p.priv_cp[p.priv_ii[i] + 1] = p.y[i];
    }
  }

  if p.loop_num {
    p.priv_lt = (qps_float_t *) malloc(p.loop_num * sizeof(qps_float_t));
    assert(p.priv_lt, "p.priv_lt");
                      
           
                               
                                       
       
           
     
                    
                                                  
                       
           
                                       
                                  
                                           
                
             
             
     
          
                    
             
             
     
                        
                                  
                                       
                               
                
              
       
           
     
                          
          
         
       
                   
            
           
       
            
                   
           
            
       
                          
                                    
                                     
                             
              
            
     
         
       
                            
                          
           
     
                      
  }

  for {
    qps_solve_inner(p);
  
if !(!p.loop_done || !p.max_done) { break }
} 

  /* retrieve values */
  /* qps_settp() should have already been called at this point */
  for i = p.num_cells; i--; {
    p.x[i] = p.priv_tp[i * 2];
    p.y[i] = p.priv_tp[i * 2 + 1];
  }
                   
                                
                                                                    
   
      

  free(p.priv_cp);
  if p.max_enable {
    free(p.priv_mxl);
  }
  if p.cog_num {
    free(p.priv_gt);
    free(p.priv_gm);
    free(p.priv_gw);
  }
  if p.loop_num {
    free(p.priv_lt);
                      
                     
      
  }

                       
  free(p.priv_pcg);
  free(p.priv_pcgt);
      
}

/**********************************************************************/

func qps_clean(p *qps_problem_t) {
  free(p.priv_tp);
  free(p.priv_ii);
  free(p.priv_cc);
  free(p.priv_cr);
  free(p.priv_cw);
  free(p.priv_ct);

                      
                     
                      
}

/**********************************************************************/
//ABC_NAMESPACE_IMPL_END

